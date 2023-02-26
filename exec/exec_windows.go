package exec

import (
	"os"

	"golang.org/x/sys/windows"
)

const (
	accessRight_PROCESS_SET_QUOTA = 0x0100
	accessRight_PROCESS_TERMINATE = 0x0001
)

func (c *Cmd) start() error {
	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.ProcessGroup = newProcessGroup(c.cmd.Process)
	if err := assignProcessToJobObject(c.ProcessGroup); err != nil {
		c.ProcessGroup.err = err
	}

	return nil
}

func assignProcessToJobObject(pg *ProcessGroup) error {
	procHandle, err := windows.OpenProcess(accessRight_PROCESS_SET_QUOTA|accessRight_PROCESS_TERMINATE, false, uint32(pg.parentProcess.Pid))
	if err != nil {
		return os.NewSyscallError("OpenProcess", err)
	}
	defer windows.CloseHandle(procHandle)

	jobHandle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return os.NewSyscallError("CreateJobObject", err)
	}
	pg.jobHandle = uintptr(jobHandle)

	err = windows.AssignProcessToJobObject(jobHandle, procHandle)
	if err != nil {
		pg.release()
		return os.NewSyscallError("AssignProcessToJobObject", err)
	}

	return nil
}
