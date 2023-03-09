package gexec

import (
	"os"
	"os/exec"
)

type GroupedCmd struct {
	*exec.Cmd
	pgid      int
	jobObject *jobObject
}

func Grouped(cmd *exec.Cmd) *GroupedCmd {
	return &GroupedCmd{Cmd: cmd}
}

func (c *GroupedCmd) Start() error {
	return c.start()
}

func (c *GroupedCmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

func (c *GroupedCmd) SignalAll(sig os.Signal) error {
	return c.signalAll(sig)
}

func (c *GroupedCmd) KillAll() error {
	return c.SignalAll(os.Kill)
}
