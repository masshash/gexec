//go:build windows
// +build windows

package gexec

import (
	"errors"
	"os"
	"runtime"

	"golang.org/x/sys/windows"
)

const NULL = 0

func newJobObject() *jobObject {
	job := &jobObject{handle: NULL}
	runtime.SetFinalizer(job, (*jobObject).close)
	return job
}

func checkValidJobHandle(jobHandle windows.Handle) error {
	if jobHandle == windows.InvalidHandle {
		return errors.New("job object already closed")
	}
	if jobHandle == NULL {
		return errors.New("job object not initialized")
	}
	return nil
}

func (job *jobObject) close() error {
	job.sigMu.Lock()
	defer job.sigMu.Unlock()

	jobHandle := windows.Handle(job.handle)
	if err := checkValidJobHandle(jobHandle); err != nil {
		return err
	}

	if err := windows.CloseHandle(jobHandle); err != nil {
		return os.NewSyscallError("CloseHandle", err)
	}
	job.handle = uintptr(windows.InvalidHandle)
	runtime.SetFinalizer(job, nil)
	return nil
}

func (job *jobObject) terminate() error {
	job.sigMu.RLock()
	defer job.sigMu.RUnlock()

	jobHandle := windows.Handle(job.handle)
	if err := checkValidJobHandle(jobHandle); err != nil {
		return err
	}
	if err := windows.TerminateJobObject(jobHandle, 1); err != nil {
		return os.NewSyscallError("TerminateJobObject", err)
	}
	return nil
}
