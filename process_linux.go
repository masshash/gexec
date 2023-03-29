//go:build linux
// +build linux

package gexec

import (
	"errors"
	"syscall"

	"golang.org/x/sys/unix"
)

func (p *Process) tryWait() error {
	_, err := p.Wait()
	if err != nil && !errors.Is(err, syscall.ECHILD) {
		return err
	}

	if err := unix.PtraceSeize(p.Pid); err != nil {
		return err
	}

	return unix.Waitid(unix.P_PID, p.Pid, &unix.Siginfo{}, unix.WEXITED|unix.WNOWAIT, nil)
}
