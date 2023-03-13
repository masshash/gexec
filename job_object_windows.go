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

// process-specific access rights.
// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const (
	PROCESS_SET_QUOTA = 0x0100
	PROCESS_TERMINATE = 0x0001
)

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
		return errors.New("job object not created")
	}
	return nil
}

func (job *jobObject) assignProcess(process *os.Process) {
	if job == nil {
		return
	}

	procHandle, err := windows.OpenProcess(PROCESS_SET_QUOTA|PROCESS_TERMINATE, false, uint32(process.Pid))
	if err != nil {
		job.Err = os.NewSyscallError("OpenProcess", err)
		return
	}
	defer windows.CloseHandle(procHandle)

	jobHandle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		job.Err = os.NewSyscallError("CreateJobObject", err)
		return
	}
	job.handle = uintptr(jobHandle)

	err = windows.AssignProcessToJobObject(jobHandle, procHandle)
	if err != nil {
		defer job.close()
		job.Err = os.NewSyscallError("AssignProcessToJobObject", err)
	}
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
