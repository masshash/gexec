package exec

import (
	"os/exec"
)

type Cmd struct {
	Cmd          *exec.Cmd
	ProcessGroup *ProcessGroup
}

func (c *Cmd) Start() error {
	return c.start()
}
