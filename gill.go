package gill

import (
	"os/exec"

	gillExec "github.com/masshash/gill/exec"
)

func Command(name string, arg ...string) *gillExec.Cmd {
	c := exec.Command(name, arg...)
	return &gillExec.Cmd{Cmd: c}
}
