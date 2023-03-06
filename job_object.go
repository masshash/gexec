package gexec

import (
	"sync"
)

type jobObject struct {
	handle uintptr
	err    error
	sigMu  sync.RWMutex
}
