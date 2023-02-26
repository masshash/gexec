package exec

import (
	"errors"
	"os"
	"runtime"

	gillError "github.com/masshash/gill/error"

	"golang.org/x/sys/windows"
)

func checkValidJobHandle(jobHandle windows.Handle) error {
	if jobHandle == windows.InvalidHandle {
		return errors.New(gillError.PROCESSGROUP_RELEASED)
	}
	if jobHandle == NULL {
		return errors.New(gillError.PROCESSGROUP_NOT_INITIALIZED)
	}
	return nil
}

func (pg *ProcessGroup) release() error {
	pg.sigMu.Lock()
	defer pg.sigMu.Unlock()

	jobHandle := windows.Handle(pg.jobHandle)
	if err := checkValidJobHandle(jobHandle); err != nil {
		return err
	}
	if err := windows.CloseHandle(jobHandle); err != nil {
		return os.NewSyscallError("CloseHandle", err)
	}
	pg.jobHandle = uintptr(windows.InvalidHandle)
	runtime.SetFinalizer(pg, nil)
	return nil
}

func (pg *ProcessGroup) signal(sig os.Signal) error {
	pg.sigMu.RLock()
	defer pg.sigMu.RUnlock()

	jobHandle := windows.Handle(pg.jobHandle)
	if err := checkValidJobHandle(jobHandle); err != nil {
		return err
	}
	if sig != os.Kill {
		return errors.New(gillError.UNSUPPORTED_SIGNAL)
	}
	if err := windows.TerminateJobObject(jobHandle, 1); err != nil {
		return os.NewSyscallError("TerminateJobObject", err)
	}
	return nil
}
