package gexec

import (
	"sync"
)

type jobObject struct {
	handle uintptr
	Err    error
	sigMu  sync.RWMutex
}
