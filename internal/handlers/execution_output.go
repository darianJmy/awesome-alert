package handlers

import (
	"bytes"
	"sync"
)

// synchronizedBuffer preserves the terminal-style combined output while
// allowing stdout and stderr copy loops to write concurrently.
type synchronizedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	limit  int // Zero keeps the existing unlimited behavior for other callers.
}

func (buffer *synchronizedBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	n := len(value)
	if buffer.limit > 0 {
		remaining := max(0, buffer.limit-buffer.buffer.Len())
		value = value[:min(len(value), remaining)]
	}
	_, err := buffer.buffer.Write(value)
	return n, err
}

func (buffer *synchronizedBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.String()
}
