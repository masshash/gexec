//go:build linux
// +build linux

package gexec

import (
	"errors"
	"os"
	"syscall"
)

func (c *GroupedCmd) start() error {
	if c.Cmd.SysProcAttr != nil {
		c.Cmd.SysProcAttr.Setpgid = true
	} else {
		c.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := c.Cmd.Start(); err != nil {
		return err
	}
	c.pgid = c.Cmd.Process.Pid
	return nil
}

func (c *GroupedCmd) signalAll(sig os.Signal) error {
	if c.pgid == -1 {
		return errors.New("invalid process group id")
	}
	if c.pgid == 0 {
		return errors.New("process group id not assigned")
	}
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("unsupported signal type")
	}
	return syscall.Kill(-c.pgid, s)
}
