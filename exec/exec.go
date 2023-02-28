package exec

import (
	"os/exec"
)

type CmdWrapper struct {
	cmd          *exec.Cmd
	cancel       func() error
	ProcessGroup *ProcessGroup
}

func (cw *CmdWrapper) Start() error {
	return cw.start()
}

func (cw *CmdWrapper) Wait() error {
	return cw.cmd.Wait()
}

func (cw *CmdWrapper) SetCancelHandler(handler func() error) {
	cw.cmd.Cancel = handler
}

func NewCmdWrapper(cmd *exec.Cmd) *CmdWrapper {
	return &CmdWrapper{cmd: cmd}
}
