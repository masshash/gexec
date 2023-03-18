//go:build windows
// +build windows

package gexec

import (
	"errors"
	"os"
	"reflect"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const NULL = 0

// process-specific access rights.
// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const (
	PROCESS_SET_QUOTA = 0x0100
	PROCESS_TERMINATE = 0x0001
)

const JobObjectBasicProcessIdList = 3

type JOBOBJECT_BASIC_PROCESS_ID_LIST struct {
	NumberOfAssignedProcesses uint32
	NumberOfProcessIdsInList  uint32
	ProcessIdList             [1]uintptr
}

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

func (job *jobObject) processes() ([]*os.Process, error) {
	job.sigMu.RLock()
	defer job.sigMu.RUnlock()

	jobHandle := windows.Handle(job.handle)
	if err := checkValidJobHandle(jobHandle); err != nil {
		return nil, err
	}

	info := JOBOBJECT_BASIC_PROCESS_ID_LIST{}
	err := windows.QueryInformationJobObject(
		jobHandle,
		JobObjectBasicProcessIdList,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
		nil,
	)

	// This is either the case where there is only one process or no processes in
	// the job. Any other case will result in ERROR_MORE_DATA. Check if info.NumberOfProcessIdsInList
	// is 1 and just return this, otherwise return an empty slice.
	if err == nil {
		if info.NumberOfProcessIdsInList == 1 {
			if p, err := os.FindProcess(int(info.ProcessIdList[0])); err == nil {
				return []*os.Process{p}, nil
			}
		}
		// Return empty slice instead of nil to play well with the caller of this.
		// Do not return an error if no processes are running inside the job
		return []*os.Process{}, nil
	}

	if err != windows.ERROR_MORE_DATA {
		return nil, os.NewSyscallError("QueryInformationJobObject", err)
	}

	adjustedInfoSize := unsafe.Sizeof(info) + (unsafe.Sizeof(info.ProcessIdList[0]) * uintptr(info.NumberOfAssignedProcesses-1))
	buf := make([]byte, adjustedInfoSize)
	if err = windows.QueryInformationJobObject(
		jobHandle,
		JobObjectBasicProcessIdList,
		uintptr(unsafe.Pointer(&buf[0])),
		uint32(len(buf)),
		nil,
	); err != nil {
		return nil, os.NewSyscallError("QueryInformationJobObject", err)
	}

	processIdList := make([]uintptr, info.NumberOfAssignedProcesses)
	dataOffset := int(unsafe.Sizeof(info.NumberOfAssignedProcesses) + unsafe.Sizeof(info.NumberOfProcessIdsInList))
	(*reflect.SliceHeader)(unsafe.Pointer(&processIdList)).Data = uintptr(unsafe.Pointer(&buf[dataOffset]))

	processes := []*os.Process{}
	for _, pid := range processIdList {
		if p, err := os.FindProcess(int(pid)); err == nil {
			processes = append(processes, p)
		}
	}

	return processes, nil
}
