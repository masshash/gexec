//go:build !windows && !plan9 && !js && !wasm

package exec

import (
	"syscall"
)

func (c *Cmd) start() error {
	if c.Cmd.SysProcAttr != nil {
		c.Cmd.SysProcAttr.Setpgid = true
	} else {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := c.Cmd.Start(); err != nil {
		return err
	}

	c.ProcessGroup = newProcessGroup(c.Cmd.Process)
	return nil
}
