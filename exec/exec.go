package exec

import (
	"os/exec"
)

type Cmd struct {
	cmd          *exec.Cmd
	cancel       func() error
	ProcessGroup *ProcessGroup
}

func (c *Cmd) Start() error {
	return c.start()
}

func (c *Cmd) Wait() error {
	return c.cmd.Wait()
}

func (c *Cmd) SetCancelHandler(handler func() error) {
	c.cmd.Cancel = handler
}

func NewCmd(cmd *exec.Cmd) *Cmd {
	return &Cmd{cmd: cmd}
}
