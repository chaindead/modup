package govulncheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/vuln/scan"
)

func GetVunerabilities() ([]Vunerability, error) {
	ctx := context.Background()
	cmd := scan.Command(ctx, []string{"-json", "./..."}...)

	// stream stderr directly
	cmd.Stderr = os.Stderr

	// wire stdout to a pipe we can consume as a stream
	pr, pw := io.Pipe()
	cmd.Stdout = pw

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("govulncheck start failed: %w", err)
	}

	collector := newCollector()
	var parseErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		parseErr = handleJSON(pr, collector)
	}()

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("govulncheck wait failed: %w", err)
	}
	_ = pw.Close()
	<-done
	if parseErr != nil && parseErr != io.EOF {
		return nil, parseErr
	}

	return collector.toVulnerabilities(), nil
}

type Vunerability struct {
	Description  string
	URL          string
	Module       string
	Version      string
	FixedVersion string
	Examples     []string
}

type collector struct {
	mu       sync.Mutex
	entries  map[string]*Entry
	findings map[string][]*Finding
}

func newCollector() *collector {
	return &collector{
		entries:  make(map[string]*Entry),
		findings: make(map[string][]*Finding),
	}
}

// gc.Handler impl
func (c *collector) Config(_ *Config) error     { return nil }
func (c *collector) SBOM(_ *SBOM) error         { return nil }
func (c *collector) Progress(_ *Progress) error { return nil }

func (c *collector) OSV(entry *Entry) error {
	c.mu.Lock()
	c.entries[entry.ID] = entry
	c.mu.Unlock()
	return nil
}

func (c *collector) Finding(f *Finding) error {
	c.mu.Lock()
	c.findings[f.OSV] = append(c.findings[f.OSV], f)
	c.mu.Unlock()
	return nil
}

func (c *collector) toVulnerabilities() []Vunerability {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result []Vunerability
	for id, flist := range c.findings {
		entry := c.entries[id]
		if entry == nil {
			continue
		}

		// group by module@version from the first frame
		type mvKey struct{ module, version string }
		byMV := make(map[mvKey][]*Finding)
		for _, f := range flist {
			k := mvKey{}
			if len(f.Trace) > 0 && f.Trace[0] != nil {
				k.module = f.Trace[0].Module
				k.version = f.Trace[0].Version
			}
			byMV[k] = append(byMV[k], f)
		}

		for k, group := range byMV {
			if len(group) == 0 {
				continue
			}

			description := entry.Summary
			if strings.TrimSpace(description) == "" {
				description = entry.Details
			}
			url := ""
			if entry.DatabaseSpecific != nil {
				url = entry.DatabaseSpecific.URL
			}

			var examples []string
			for _, f := range group {
				if !isTraceExampleEligible(f.Trace) {
					continue
				}
				if ex := formatTraceExample(f.Trace); ex != "" {
					examples = append(examples, ex)
				}
			}

			fixed := group[0].FixedVersion

			result = append(result, Vunerability{
				Description:  description,
				URL:          url,
				Module:       k.module,
				Version:      k.version,
				FixedVersion: fixed,
				Examples:     dedupStrings(examples),
			})
		}
	}
	return result
}

func formatTraceExample(trace []*Frame) string {
	pos := lastPosition(trace)
	chain := formatCallChain(trace)
	if pos == "" {
		return chain
	}
	return pos + ": " + chain
}

func dedupStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func lastPosition(trace []*Frame) string {
	// Choose the deepest frame that has a position; fallback to first available.
	for i := len(trace) - 1; i >= 0; i-- {
		if trace[i] != nil && trace[i].Position != nil && trace[i].Position.Filename != "" {
			p := trace[i].Position
			return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
		}
	}
	// If none, return empty to skip placeholder entries
	return ""
}

func isTraceExampleEligible(trace []*Frame) bool {
	if len(trace) < 2 {
		return false
	}
	return lastPosition(trace) != ""
}

func formatCallChain(trace []*Frame) string {
	if len(trace) == 0 {
		return ""
	}
	// Reverse frames: entry -> vulnerable symbol
	rev := make([]*Frame, 0, len(trace))
	for i := len(trace) - 1; i >= 0; i-- {
		rev = append(rev, trace[i])
	}

	first := shortFunc(rev[0])
	if len(rev) == 1 {
		return first
	}
	second := shortFunc(rev[1])
	if len(rev) == 2 {
		return fmt.Sprintf("%s calls %s", first, second)
	}
	last := shortFunc(rev[len(rev)-1])
	return fmt.Sprintf("%s calls %s, which eventually calls %s", first, second, last)
}

func shortFunc(f *Frame) string {
	if f == nil {
		return "(unknown)"
	}
	pkg := shortPkg(f.Package)
	name := f.Function
	if f.Receiver != "" {
		if name != "" {
			name = f.Receiver + "." + name
		} else {
			name = f.Receiver
		}
	}
	if pkg == "" {
		if name != "" {
			return name
		}
		return f.Module
	}
	if name == "" {
		return pkg
	}
	return pkg + "." + name
}

func shortPkg(path string) string {
	if path == "" {
		return ""
	}
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// handleJSON decodes the stream of govulncheck messages and feeds the collector.
func handleJSON(from io.Reader, c *collector) error {
	dec := json.NewDecoder(from)
	for dec.More() {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			return err
		}
		if msg.Config != nil {
			if err := c.Config(msg.Config); err != nil {
				return err
			}
		}
		if msg.Progress != nil {
			if err := c.Progress(msg.Progress); err != nil {
				return err
			}
		}
		if msg.SBOM != nil {
			if err := c.SBOM(msg.SBOM); err != nil {
				return err
			}
		}
		if msg.OSV != nil {
			if err := c.OSV(msg.OSV); err != nil {
				return err
			}
		}
		if msg.Finding != nil {
			if err := c.Finding(msg.Finding); err != nil {
				return err
			}
		}
	}
	return nil
}
