package deps

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/mod/modfile"
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

// ListAllModulePaths returns all direct (non-indirect) module paths declared in go.mod
func ListAllModulePaths() ([]string, error) {
	gomodPath, err := getGoModPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return nil, err
	}

	f, err := modfile.Parse(gomodPath, data, nil)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(f.Require))
	seen := make(map[string]struct{})
	for _, req := range f.Require {
		if req == nil || req.Mod.Path == "" || req.Indirect {
			continue
		}
		if _, ok := seen[req.Mod.Path]; ok {
			continue
		}
		paths = append(paths, req.Mod.Path)
		seen[req.Mod.Path] = struct{}{}
	}

	return paths, nil
}

func getGoModPath() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(string(out))
	if path == "" || path == os.DevNull {
		return "go.mod", nil
	}
	return path, nil
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
