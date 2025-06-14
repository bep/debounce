// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
// Copyright © 2025 Vsevolod Strukchinsky <floatdrop@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple, thread-safe debounce library for Go that delays function execution until
// after a specified duration has elapsed since the last invocation.
// Perfect for rate limiting, reducing redundant operations, and
// optimizing performance in high-frequency scenarios.
package debounce

import (
	"math"
	"sync"
	"time"
)

// Option is a functional option for configuring the debouncer.
type Option func(*debouncer)

// WithMaxCalls sets the maximum number of calls before the debounced function is executed.
// To set no limit, use -1. By default, there is no limit.
func WithMaxCalls(count int) Option {
	return func(d *debouncer) {
		d.callsLimit = count
	}
}

// WithMaxWait sets the maximum wait time before the debounced function is executed.
func WithMaxWait(limit time.Duration) Option {
	return func(d *debouncer) {
		d.waitLimit = limit
	}
}

// New returns a debounced function that takes another functions as its argument.
// This function will be called when the debounced function stops being called
// for the given duration.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func New(after time.Duration, options ...Option) func(f func()) {
	d := &debouncer{
		after:      after,
		startWait:  time.Now(),
		waitLimit:  math.MaxInt64, // effectively no limit
		callsLimit: -1,
	}

	for _, opt := range options {
		opt(d)
	}

	return func(f func()) {
		d.add(f)
	}
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer

	calls      int
	callsLimit int

	startWait time.Time
	waitLimit time.Duration
}

func (d *debouncer) callLimitReached() bool {
	return d.callsLimit != -1 && d.calls >= d.callsLimit
}

func (d *debouncer) timeLimitReached() bool {
	return time.Since(d.startWait) >= d.waitLimit
}

func (d *debouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.calls += 1
		d.timer.Stop()
	} else {
		d.calls = 1
	}

	// If the function has been called more than the limit, or if the wait time
	// has exceeded the limit, execute the function immediately.
	if d.callLimitReached() || d.timeLimitReached() {
		d.calls = 0
		d.startWait = time.Now()
		f()
	} else { // Otherwise, set a timer to call the function after the specified duration.
		d.timer = time.AfterFunc(d.after, f)
	}
}
