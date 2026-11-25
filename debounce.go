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

// NewWithTimeout returns a debounced function that takes another function as
// its argument, like New, and additionally guarantees that the function is
// called at most timeout after the first call in a burst, even if calls keep
// arriving more often than after.
//
// This is useful when the input may never pause long enough for a plain
// debounce to fire, for example when batching a stream of database queries.
// A timeout less than or equal to zero disables the maximum wait, making the
// behavior identical to New.
func NewWithTimeout(after, timeout time.Duration) func(f func()) {
	d := &timeoutDebouncer{after: after, timeout: timeout}

	return func(f func()) {
		d.add(f)
	}
}

type timeoutDebouncer struct {
	mu       sync.Mutex
	after    time.Duration
	timeout  time.Duration
	timer    *time.Timer
	deadline time.Time
	f        func()
	gen      uint64
}

func (d *timeoutDebouncer) add(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	// A zero or expired deadline starts a new burst. Keeping the existing
	// deadline otherwise is what bounds the wait for a continuously called
	// function.
	if d.timeout > 0 && (d.deadline.IsZero() || !now.Before(d.deadline)) {
		d.deadline = now.Add(d.timeout)
	}

	delay := d.after
	if d.timeout > 0 {
		if remaining := d.deadline.Sub(now); remaining < delay {
			delay = remaining
		}
	}
	if delay < 0 {
		delay = 0
	}

	d.f = f
	d.gen++
	gen := d.gen
	d.timer = time.AfterFunc(delay, func() { d.fire(gen) })
}

func (d *timeoutDebouncer) fire(gen uint64) {
	d.mu.Lock()
	// A newer call rescheduled the timer after this one fired, so ignore the
	// stale callback.
	if gen != d.gen {
		d.mu.Unlock()
		return
	}
	d.timer = nil
	d.deadline = time.Time{}
	f := d.f
	d.f = nil
	d.mu.Unlock()

	if f != nil {
		f()
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
