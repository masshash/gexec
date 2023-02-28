package exec

import (
	"os"

	"golang.org/x/sys/windows"
)

const (
	accessRight_PROCESS_SET_QUOTA = 0x0100
	accessRight_PROCESS_TERMINATE = 0x0001
)

func (cw *CmdWrapper) start() error {
	if err := cw.cmd.Start(); err != nil {
		return err
	}

	cw.ProcessGroup = newProcessGroup(cw.cmd.Process)
	if err := assignProcessToJobObject(cw.ProcessGroup); err != nil {
		cw.ProcessGroup.err = err
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
