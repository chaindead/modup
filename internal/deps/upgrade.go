package deps

import (
	"fmt"
	"os"
	"os/exec"
)

func Upgrade(m Module) error {
	target := "v" + m.Latest.String()

	cmd := exec.Command("go", "get", fmt.Sprintf("%s@%s", m.Path, target))
	cmd.Env = append(os.Environ(), "GOWORK=off")

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go get %s failed: %v\n%s", m.Path, err, string(out))
	}

	return nil
}
