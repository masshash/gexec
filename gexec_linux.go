//go:build linux
// +build linux

package gexec

import (
	"errors"
	"os"
	"syscall"
)

func (gc *GroupedCmd) start() error {
	if gc.Cmd.SysProcAttr != nil {
		gc.Cmd.SysProcAttr.Setpgid = true
	} else {
		gc.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := gc.Cmd.Start(); err != nil {
		return err
	}
	gc.pgid = gc.Cmd.Process.Pid
	return nil
}

func (gc *GroupedCmd) signalAll(sig os.Signal) error {
	if gc.pgid == -1 {
		return errors.New("invalid process group id")
	}
	if gc.pgid == 0 {
		return errors.New("process group not initialized")
	}
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("unsupported signal type")
	}
	return syscall.Kill(-gc.pgid, s)
}
