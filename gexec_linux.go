//go:build linux
// +build linux

package gexec

import (
	"errors"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
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

func checkValidPgid(pgid int) error {
	if pgid == -1 {
		return errors.New("invalid process group id")
	}
	if pgid == 0 {
		return errors.New("process group id not assigned")
	}
	return nil
}

func (c *GroupedCmd) signalAll(sig os.Signal) error {
	if err := checkValidPgid(c.pgid); err != nil {
		return err
	}

	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("unsupported signal type")
	}
	return syscall.Kill(-c.pgid, s)
}

func (c *GroupedCmd) processes() ([]*Process, error) {
	if err := checkValidPgid(c.pgid); err != nil {
		return nil, err
	}

	pids, err := readPidsFromProc()
	if err != nil {
		return nil, err
	}

	var processes []*Process
	for _, pid := range pids {
		findPgid, err := unix.Getpgid(pid)
		if err == nil && findPgid == c.pgid {
			p, _ := os.FindProcess(pid)
			processes = append(processes, &Process{Process: p})
		}
	}
	return processes, nil
}

func readPidsFromProc() ([]int, error) {
	var ret []int

	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fnames, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	for _, fname := range fnames {
		pid, err := strconv.ParseInt(fname, 10, 32)
		if err != nil {
			// if not numeric name, just skip
			continue
		}
		ret = append(ret, int(pid))
	}
	return ret, nil
}
