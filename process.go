package gexec

import "os"

type Process struct {
	*os.Process
}

func (p *Process) TryWait() error {
	return p.tryWait()
}
