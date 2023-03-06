//go:build windows
// +build windows

package gexec

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// process-specific access rights.
// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const (
	PROCESS_SET_QUOTA = 0x0100
	PROCESS_TERMINATE = 0x0001
)

func (gc *GroupedCmd) start() error {
	if err := gc.Cmd.Start(); err != nil {
		return err
	}
	gc.jobObject = newJobObject()
	assignProcessToJobObject(gc.Cmd.Process, gc.jobObject)
	return nil
}

func assignProcessToJobObject(process *os.Process, job *jobObject) {
	procHandle, err := windows.OpenProcess(PROCESS_SET_QUOTA|PROCESS_TERMINATE, false, uint32(process.Pid))
	if err != nil {
		job.err = os.NewSyscallError("OpenProcess", err)
		return
	}
	defer windows.CloseHandle(procHandle)

	jobHandle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		job.err = os.NewSyscallError("CreateJobObject", err)
		return
	}
	job.handle = uintptr(jobHandle)

	err = windows.AssignProcessToJobObject(jobHandle, procHandle)
	if err != nil {
		defer job.close()
		job.err = os.NewSyscallError("AssignProcessToJobObject", err)
	}
}

func (gc *GroupedCmd) signalAll(sig os.Signal) error {
	if sig != os.Kill {
		return errors.New("unsupported signal type")
	}

	return gc.jobObject.terminate()
}
