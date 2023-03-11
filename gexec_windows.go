//go:build windows
// +build windows

package gexec

import (
	"errors"
	"os"
)

func (c *GroupedCmd) start() error {
	if err := c.Cmd.Start(); err != nil {
		return err
	}
	c.jobObject = newJobObject()
	c.jobObject.assignProcess(c.Cmd.Process)
	return nil
}

func (c *GroupedCmd) signalAll(sig os.Signal) error {
	if sig != os.Kill {
		return errors.New("unsupported signal type")
	}

	return c.jobObject.terminate()
}
