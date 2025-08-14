package deps

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
)

// Module describes a Go module upgrade candidate
type Module struct {
	Path           string
	Current        *semver.Version
	Latest         *semver.Version
	IsTool         bool
	UpdateCategory string // "minor" | "patch" | "prerelease" | "metadata"
}

// goListModule mirrors a subset of fields from `go list -u -m -json` output
type goListModule struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
	Main     bool   `json:"Main"`
	Update   *struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Update"`
}

// DiscoverOutdated returns non-main, non-indirect modules that have updates available
func DiscoverOutdated() ([]Module, error) {
	args := []string{"list", "-u", "-m", "-json", "all"}
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list failed: %w", err)
	}
	mods, err := parseGoListJSON(string(out))
	if err != nil {
		return nil, err
	}
	return mods, nil
}

// ToolsSupported returns whether `go tool` upgrades can be inspected (Go>=1.24)
func ToolsSupported() (bool, error) {
	gv, err := exec.Command("go", "version").Output()
	if err != nil {
		return false, err
	}
	version := strings.TrimSpace(string(gv))
	re := regexp.MustCompile(`go version go([\d\.]+)(rc.+)?`)
	m := re.FindStringSubmatch(version)
	if len(m) < 2 {
		return false, fmt.Errorf("couldn't parse go version %s", version)
	}
	v, err := semver.NewVersion(m[1])
	if err != nil {
		return false, err
	}
	if v.Major() >= 1 && v.Minor() >= 24 {
		return true, nil
	}
	return false, nil
}

// DiscoverToolUpdates finds updates for modules referenced by `go tool` output
func DiscoverToolUpdates() ([]Module, error) {
	cmd := exec.Command("go", "list", "-f", "{{if .Module}}{{.Module.Path}} {{.Module.Version}}{{end}}", "tool")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	toolsOut, err := cmd.Output()
	if err != nil {
		// When there are no tool modules configured, `go list tool` may fail
		if strings.Contains(string(err.Error()), "matched no packages") {
			return []Module{}, nil
		}
		return nil, fmt.Errorf("listing tools failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(toolsOut)), "\n")
	var result []Module
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		path := parts[0]
		current := parts[1]
		updCmd := exec.Command("go", "list", "-m", "-f", "{{if .Update}}{{.Update.Version}}{{end}}", "-u", path)
		updCmd.Env = append(os.Environ(), "GOWORK=off")
		latestOut, err := updCmd.Output()
		if err != nil {
			continue
		}
		latest := strings.TrimSpace(string(latestOut))
		if latest == "" || latest == current {
			continue
		}
		fromV, err := semver.NewVersion(stripV(current))
		if err != nil {
			continue
		}
		toV, err := semver.NewVersion(stripV(latest))
		if err != nil {
			continue
		}
		result = append(result, Module{
			Path:           path,
			Current:        fromV,
			Latest:         toV,
			IsTool:         true,
			UpdateCategory: categorize(fromV, toV),
		})
	}
	return result, nil
}

// CountModules returns total number of modules in the current workspace/module
func CountModules() (int, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Path}}", "all")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("count modules failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	if count == 0 {
		return 0, fmt.Errorf("no modules found")
	}
	return count, nil
}

// ListAllModulePaths returns all module paths in the current context
func ListAllModulePaths() ([]string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Path}}", "all")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var paths []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			paths = append(paths, l)
		}
	}
	return paths, nil
}

// DiscoverOutdatedWithProgress discovers outdated modules and reports progress per processed module
func DiscoverOutdatedWithProgress(progress chan<- int) ([]Module, error) {
	paths, err := ListAllModulePaths()
	if err != nil {
		return nil, err
	}
	type res struct {
		m  Module
		ok bool
	}
	results := make(chan res, len(paths))
	// limit concurrency to avoid overwhelming network
	const maxWorkers = 8
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	for _, p := range paths {
		path := p
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			cmd := exec.Command("go", "list", "-m", "-u", "-json", path)
			cmd.Env = append(os.Environ(), "GOWORK=off")
			out, err := cmd.Output()
			if progress != nil {
				select {
				case progress <- 1:
				default:
				}
			}
			if err != nil {
				results <- res{ok: false}
				return
			}
			var m goListModule
			if e := json.Unmarshal(out, &m); e != nil {
				results <- res{ok: false}
				return
			}
			if m.Main || m.Indirect || m.Update == nil || m.Update.Version == "" {
				results <- res{ok: false}
				return
			}
			fromV, e1 := semver.NewVersion(stripV(m.Version))
			toV, e2 := semver.NewVersion(stripV(m.Update.Version))
			if e1 != nil || e2 != nil {
				results <- res{ok: false}
				return
			}
			results <- res{m: Module{Path: m.Path, Current: fromV, Latest: toV, UpdateCategory: categorize(fromV, toV)}, ok: true}
		}()
	}
	go func() { wg.Wait(); close(results) }()
	var modules []Module
	for r := range results {
		if r.ok {
			modules = append(modules, r.m)
		}
	}
	return modules, nil
}

func parseGoListJSON(data string) ([]Module, error) {
	dec := json.NewDecoder(bufio.NewReader(strings.NewReader(data)))
	var modules []Module
	for {
		var m goListModule
		if err := dec.Decode(&m); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// The output may contain concatenated JSON objects without commas; try to recover
			return nil, fmt.Errorf("parse go list json: %w", err)
		}
		if m.Main || m.Indirect {
			continue
		}
		if m.Update == nil || m.Update.Version == "" {
			continue
		}
		fromV, err := semver.NewVersion(stripV(m.Version))
		if err != nil {
			continue
		}
		toV, err := semver.NewVersion(stripV(m.Update.Version))
		if err != nil {
			continue
		}
		modules = append(modules, Module{
			Path:           m.Path,
			Current:        fromV,
			Latest:         toV,
			UpdateCategory: categorize(fromV, toV),
		})
	}
	return modules, nil
}
func stripV(v string) string {
	return strings.TrimPrefix(v, "v")
}

func categorize(from, to *semver.Version) string {
	if from.Major() != to.Major() {
		return "major"
	}
	if from.Minor() != to.Minor() {
		return "minor"
	}
	if from.Patch() != to.Patch() {
		return "patch"
	}
	if to.Prerelease() != "" && from.Prerelease() != to.Prerelease() {
		return "prerelease"
	}
	if to.Metadata() != "" && from.Metadata() != to.Metadata() {
		return "metadata"
	}
	return "same"
}
