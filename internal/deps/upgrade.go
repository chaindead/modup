package deps

import (
	"fmt"
	"os"
	"os/exec"
)

func ApplyProgressDetailed(modules []Module, progress chan<- Module) error {
	for _, m := range modules {
		target := "v" + m.Latest.String()

		cmd := exec.Command("go", "get", fmt.Sprintf("%s@%s", m.Path, target))
		cmd.Env = append(os.Environ(), "GOWORK=off")

		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go get %s failed: %v\n%s", m.Path, err, string(out))
		}

		if progress != nil {
			select {
			case progress <- m:
			default:
			}
		}
	}

	return nil
}
