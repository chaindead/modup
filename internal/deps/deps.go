package deps

import (
	"encoding/json"
	"os"
	"os/exec"
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
