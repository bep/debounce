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

// New wraps a function and returns a wrapped debounce function, that,
// as long as it continues to be invoked, will not be triggered.
// The function will be called after it stops being called for N milliseconds
// If `immediate` is passed, trigger the function on the leading edge, instead of the trailing.
func New(after time.Duration, immediate bool) (debounce func(f func())) {
	d := &debouncer{after: after}

	if !immediate {
		return func(f func()) {
			d.debounce(f)
		}
	}
	return func(f func()) {
		d.debounced(f)
	}
}

type debouncer struct {
	mu    sync.Mutex
	after time.Duration
	timer *time.Timer
}

func (d *debouncer) debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.after, f)
}

func (d *debouncer) debounced(f func()) {
	d.mu.Lock()

	if d.timer == nil {
		f()
		d.timer = time.AfterFunc(d.after, func() {
			d.mu.Lock()
			d.timer.Stop()
			d.timer = nil
			d.mu.Unlock()
		})
	}
	d.timer.Stop()
	d.timer = time.AfterFunc(d.after, func() {
		d.mu.Lock()
		d.timer.Stop()
		d.timer = nil
		d.mu.Unlock()
	})

	d.mu.Unlock()
}
