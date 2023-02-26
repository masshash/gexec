package exec

import (
	"os/exec"
)

type Cmd struct {
	cmd          *exec.Cmd
	ProcessGroup *ProcessGroup
}

func (c *Cmd) Start() error {
	return c.start()
}

func NewCmd(cmd *exec.Cmd) *Cmd {
	return &Cmd{cmd: cmd, ProcessGroup: &ProcessGroup{}}
}
