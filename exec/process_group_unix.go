//go:build !windows && !plan9 && !js && !wasm

package exec

import (
	"errors"
	"os"
	"syscall"

	gillError "github.com/masshash/gill/error"
)

func checkValidPgid(pgid int) error {
	if pgid == -1 {
		return errors.New(gillError.PROCESSGROUP_RELEASED)
	}
	if pgid == 0 {
		return errors.New(gillError.PROCESSGROUP_NOT_INITIALIZED)
	}
	return nil
}

func (pg *ProcessGroup) release() error {
	pg.sigMu.Lock()
	defer pg.sigMu.Unlock()

	if err := checkValidPgid(pg.pgid); err != nil {
		return err
	}

	pg.pgid = -1
	return nil
}

func (pg *ProcessGroup) signal(sig os.Signal) error {
	pg.sigMu.RLock()
	defer pg.sigMu.RUnlock()

	if err := checkValidPgid(pg.pgid); err != nil {
		return err
	}

	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New(gillError.UNSUPPORTED_SIGNAL)
	}
	return syscall.Kill(-pg.pgid, s)
}
