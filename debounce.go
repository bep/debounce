// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package debounce provides a debouncer func. The most typical use case would be
// the user typing a text into a form; the UI needs an update, but let's wait for
// a break.
package debounce

import (
	"sync"
	"time"
)

// New returns a debounced function that takes another functions as its argument.
// This function will be called when the debounced function stops being called
// for the given duration.
// The debounced function can be invoked with different functions, if needed,
// the last one will win.
func New(after time.Duration) func(f func()) {
	d := &debouncer{after: after}

	return func(f func()) {
		d.add(f)
	}
}

// NewWithCancel returns a debounced function together with a cancel function.
// The debounced function behaves like the one returned by New: it takes another
// function as its argument, and that function will be invoked when calls to the
// debounced function have stopped for the given duration. If invoked multiple
// times, the last provided function will win.
//
// The returned cancel function stops any pending timer and prevents the
// currently scheduled function (if any) from being called. Calling cancel has
// no effect if no function is scheduled or if it already executed.
//
// This is useful in shutdown scenarios where the final scheduled function must
// be suppressed or handled explicitly.
func NewWithCancel(after time.Duration) (func(f func()), func()) {
	d := &debouncer{after: after}

	return func(f func()) {
		d.add(f)
	}, d.cancel
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

func (d *debouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}

func (d *debouncer) cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}
