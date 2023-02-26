//go:build !windows && !plan9 && !js && !wasm

package exec

import (
	"syscall"
)

func (c *Cmd) start() error {
	if c.cmd.SysProcAttr != nil {
		c.cmd.SysProcAttr.Setpgid = true
	} else {
		c.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.ProcessGroup = newProcessGroup(c.cmd.Process)
	return nil
}
