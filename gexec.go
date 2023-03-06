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

func (gc *GroupedCmd) Start() error {
	return gc.start()
}

func (gc *GroupedCmd) SignalAll(sig os.Signal) error {
	return gc.signalAll(sig)
}

func (gc *GroupedCmd) KillAll() error {
	return gc.SignalAll(os.Kill)
}

func Grouped(cmd *exec.Cmd) *GroupedCmd {
	return &GroupedCmd{Cmd: cmd}
}
