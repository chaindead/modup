package deps

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Module describes a Go module upgrade candidate
type Module struct {
	Path           string
	Current        *semver.Version
	Latest         *semver.Version
	IsTool         bool
	UpdateCategory string // "minor" | "patch" | "prerelease" | "metadata"
	Updatable      bool
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

func GetModuleInfo(path string) (Module, error) {
	cmd := exec.Command("go", "list", "-m", "-u", "-json", path)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.Output()
	if err != nil {
		return Module{Path: path}, err
	}

	var m goListModule
	if e := json.Unmarshal(out, &m); e != nil {
		return Module{Path: path}, e
	}
	if m.Main || m.Indirect || m.Update == nil || m.Update.Version == "" {
		return Module{Path: path}, nil
	}

	fromV, e1 := semver.NewVersion(stripV(m.Version))
	toV, e2 := semver.NewVersion(stripV(m.Update.Version))
	if e1 != nil || e2 != nil {
		return Module{Path: path}, nil
	}

	return Module{
		Path:           m.Path,
		Current:        fromV,
		Latest:         toV,
		UpdateCategory: categorize(fromV, toV),
		Updatable:      true,
	}, nil
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
