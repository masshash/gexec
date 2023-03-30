//go:build windows
// +build windows

package gexec

func (p *Process) tryWait() error {
	if _, err := p.Wait(); err != nil {
		return err
	}
	return nil
}
