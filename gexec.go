package gexec

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type GroupedCmd struct {
	*exec.Cmd
	pgid      int
	jobObject *jobObject
}

func Grouped(cmd *exec.Cmd) *GroupedCmd {
	return &GroupedCmd{Cmd: cmd}
}

func (c *GroupedCmd) Start() error {
	if c.Process != nil {
		return errors.New("already started")
	}
	return c.start()
}

func (c *GroupedCmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

func (c *GroupedCmd) SignalAll(sig os.Signal) error {
	return c.signalAll(sig)
}

func (c *GroupedCmd) KillAll() error {
	return c.SignalAll(os.Kill)
}

func (c *GroupedCmd) Pgid() int {
	return c.pgid
}

func (c *GroupedCmd) JobObject() *jobObject {
	return c.jobObject
}

func (c *GroupedCmd) Processes() ([]*Process, error) {
	return c.processes()
}

func (c *GroupedCmd) WaitAll() error {
	return c.WaitAllWithContext(context.Background(), os.Kill)
}

func signalEachProcess(processes []*Process, sig os.Signal, mu *sync.RWMutex) {
	mu.Lock()
	defer mu.Unlock()

	var wg sync.WaitGroup
	for _, p := range processes {
		wg.Add(1)
		go func(p *Process) {
			defer wg.Done()
			p.Signal(sig)
		}(p)
	}
	wg.Wait()
}

func (c *GroupedCmd) WaitAllWithContext(ctx context.Context, sig os.Signal) error {
	if ctx == nil {
		panic("nil Context")
	}
	var mainWg sync.WaitGroup
	var mu sync.RWMutex
	var processes []*Process
	var returnErr error
	waitDone := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			signalEachProcess(processes, sig, &mu)
			c.SignalAll(sig)
			<-waitDone
		case <-waitDone:
		}
	}()

	mainWg.Add(1)
	go func() {
		defer mainWg.Done()
		var subWg sync.WaitGroup
		for {
			mu.Lock()
			select {
			case <-ctx.Done():
				mu.Unlock()
				return
			default:
			}
			processes, returnErr = c.Processes()
			mu.Unlock()
			if returnErr != nil || len(processes) == 0 {
				return
			}

			var errcnt uint32
			for _, p := range processes {
				subWg.Add(1)
				go func(p *Process) {
					defer subWg.Done()
					var err error
					if p.Pid == c.Process.Pid {
						err = c.Wait()
					} else {
						err = p.TryWait()
					}
					if err != nil {
						atomic.AddUint32(&errcnt, 1)
					}
				}(p)
			}
			subWg.Wait()

			if errcnt > 0 {
				time.Sleep(time.Millisecond * 250)
			}
		}
	}()
	mainWg.Wait()
	waitDone <- struct{}{}

	return returnErr
}

// Output runs the command and returns its standard output.
// Any returned error will usually be of type *ExitError.
// If c.Stderr was nil, Output populates ExitError.Stderr.
func (c *GroupedCmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var stdout bytes.Buffer
	c.Stdout = &stdout

	captureErr := c.Stderr == nil
	if captureErr {
		c.Stderr = &prefixSuffixSaver{N: 32 << 10}
	}

	err := c.Run()
	if err != nil && captureErr {
		if ee, ok := err.(*exec.ExitError); ok {
			ee.Stderr = c.Stderr.(*prefixSuffixSaver).Bytes()
		}
	}
	return stdout.Bytes(), err
}

// CombinedOutput runs the command and returns its combined standard
// output and standard error.
func (c *GroupedCmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Run()
	return b.Bytes(), err
}

// prefixSuffixSaver is an io.Writer which retains the first N bytes
// and the last N bytes written to it. The Bytes() methods reconstructs
// it with a pretty error message.
type prefixSuffixSaver struct {
	N         int // max size of prefix or suffix
	prefix    []byte
	suffix    []byte // ring buffer once len(suffix) == N
	suffixOff int    // offset to write into suffix
	skipped   int64

	// TODO(bradfitz): we could keep one large []byte and use part of it for
	// the prefix, reserve space for the '... Omitting N bytes ...' message,
	// then the ring buffer suffix, and just rearrange the ring buffer
	// suffix when Bytes() is called, but it doesn't seem worth it for
	// now just for error messages. It's only ~64KB anyway.
}

func (w *prefixSuffixSaver) Write(p []byte) (n int, err error) {
	lenp := len(p)
	p = w.fill(&w.prefix, p)

	// Only keep the last w.N bytes of suffix data.
	if overage := len(p) - w.N; overage > 0 {
		p = p[overage:]
		w.skipped += int64(overage)
	}
	p = w.fill(&w.suffix, p)

	// w.suffix is full now if p is non-empty. Overwrite it in a circle.
	for len(p) > 0 { // 0, 1, or 2 iterations.
		n := copy(w.suffix[w.suffixOff:], p)
		p = p[n:]
		w.skipped += int64(n)
		w.suffixOff += n
		if w.suffixOff == w.N {
			w.suffixOff = 0
		}
	}
	return lenp, nil
}

// fill appends up to len(p) bytes of p to *dst, such that *dst does not
// grow larger than w.N. It returns the un-appended suffix of p.
func (w *prefixSuffixSaver) fill(dst *[]byte, p []byte) (pRemain []byte) {
	if remain := w.N - len(*dst); remain > 0 {
		add := minInt(len(p), remain)
		*dst = append(*dst, p[:add]...)
		p = p[add:]
	}
	return p
}

func (w *prefixSuffixSaver) Bytes() []byte {
	if w.suffix == nil {
		return w.prefix
	}
	if w.skipped == 0 {
		return append(w.prefix, w.suffix...)
	}
	var buf bytes.Buffer
	buf.Grow(len(w.prefix) + len(w.suffix) + 50)
	buf.Write(w.prefix)
	buf.WriteString("\n... omitting ")
	buf.WriteString(strconv.FormatInt(w.skipped, 10))
	buf.WriteString(" bytes ...\n")
	buf.Write(w.suffix[w.suffixOff:])
	buf.Write(w.suffix[:w.suffixOff])
	return buf.Bytes()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
