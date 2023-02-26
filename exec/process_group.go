package exec

import (
	"os"
	"runtime"
	"sync"
)

const NULL = 0

type ProcessGroup struct {
	parentProcess *os.Process
	pgid          int
	jobHandle     uintptr
	err           error
	sigMu         sync.RWMutex
}

func (pg *ProcessGroup) Release() error {
	pg.parentProcess.Release()
	return pg.release()
}

func (pg *ProcessGroup) Signal(sig os.Signal) error {
	return pg.signal(sig)
}

func (pg *ProcessGroup) Kill() error {
	return pg.Signal(os.Kill)
}

func (pg *ProcessGroup) Error() error {
	return pg.err
}

func newProcessGroup(process *os.Process) *ProcessGroup {
	pg := &ProcessGroup{parentProcess: process, pgid: process.Pid, jobHandle: NULL}
	runtime.SetFinalizer(pg, (*ProcessGroup).Release)
	return pg
}
