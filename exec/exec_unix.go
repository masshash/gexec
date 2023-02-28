//go:build !windows && !plan9 && !js && !wasm

package exec

import (
	"syscall"
)

func (cw *CmdWrapper) start() error {
	if cw.cmd.SysProcAttr != nil {
		cw.cmd.SysProcAttr.Setpgid = true
	} else {
		cw.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := cw.cmd.Start(); err != nil {
		return err
	}

	cw.ProcessGroup = newProcessGroup(cw.cmd.Process)
	return nil
}
