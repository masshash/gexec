package gill

import (
	"os/exec"

	gillExec "github.com/masshash/gill/exec"
)

func Command(name string, arg ...string) *gillExec.Cmd {
	cmd := exec.Command(name, arg...)
	return gillExec.NewCmd(cmd)
}
