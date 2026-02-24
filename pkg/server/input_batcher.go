package server

import (
	"sync"
	"time"
)

// InputBatcher batches mouse move events to reduce CGo overhead.
type InputBatcher struct {
	mu          sync.Mutex
	lastMouseX  int
	lastMouseY  int
	hasPending  bool
	flushTimer  *time.Timer
	flushFunc   func(x, y int)
	batchWindow time.Duration
	stopped     bool
}

// NewInputBatcher creates a new input batcher.
func NewInputBatcher(flushFunc func(x, y int)) *InputBatcher {
	return &InputBatcher{
		flushFunc:   flushFunc,
		batchWindow: 5 * time.Millisecond,
	}
}

// AddMouseMove batches a mouse move event.
func (b *InputBatcher) AddMouseMove(x, y int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return
	}

	b.lastMouseX = x
	b.lastMouseY = y
	b.hasPending = true

	if b.flushTimer == nil {
		b.flushTimer = time.AfterFunc(b.batchWindow, b.flush)
	}
}

// Flush immediately sends pending mouse move.
func (b *InputBatcher) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.flushTimer != nil {
		b.flushTimer.Stop()
		b.flushTimer = nil
	}

	if b.hasPending {
		b.flushFunc(b.lastMouseX, b.lastMouseY)
		b.hasPending = false
	}
}

func (b *InputBatcher) flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return
	}

	if b.hasPending {
		b.flushFunc(b.lastMouseX, b.lastMouseY)
		b.hasPending = false
	}
	b.flushTimer = nil
}

// Stop stops the batcher and flushes pending events.
func (b *InputBatcher) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stopped = true
	if b.flushTimer != nil {
		b.flushTimer.Stop()
		b.flushTimer = nil
	}
	b.hasPending = false
}
