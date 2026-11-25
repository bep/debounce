// Copyright 2024 Bjørn Erik Pedersen
// SPDX-License-Identifier: MIT

package debounce_test

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bep/debounce"
)

func TestDebounceTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			calls   atomic.Int32
			firedAt time.Duration
			start   = time.Now()
		)

		debounced := debounce.NewWithTimeout(100*time.Millisecond, 250*time.Millisecond)
		record := func() {
			calls.Add(1)
			firedAt = time.Since(start)
		}

		// Keep calling more often than the debounce interval. Without a
		// timeout the function would never run.
		for range 5 {
			debounced(record)
			time.Sleep(50 * time.Millisecond)
		}

		// The call at 200ms leaves 50ms before the 250ms deadline, so it runs
		// at the deadline rather than 100ms later.
		time.Sleep(60 * time.Millisecond)

		if got := calls.Load(); got != 1 {
			t.Fatalf("expected the function to run once at the timeout, got %d", got)
		}
		if firedAt != 250*time.Millisecond {
			t.Fatalf("expected the function to run at 250ms, ran at %v", firedAt)
		}

		// After the timeout fires, a new burst gets a fresh deadline.
		debounced(record)
		time.Sleep(110 * time.Millisecond)

		if got := calls.Load(); got != 2 {
			t.Fatalf("expected a second run after the next burst, got %d", got)
		}
	})
}

func TestDebounceTimeoutTrailingCall(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32

		debounced := debounce.NewWithTimeout(100*time.Millisecond, time.Second)

		debounced(func() { calls.Add(1) })
		time.Sleep(40 * time.Millisecond)
		debounced(func() { calls.Add(2) })

		// A single burst shorter than the timeout still waits for the normal
		// debounce interval, and only the last function runs.
		time.Sleep(110 * time.Millisecond)

		if got := calls.Load(); got != 2 {
			t.Fatalf("expected only the last function to run, got marker %d", got)
		}
	})
}

func TestDebounceTimeoutDisabled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32

		debounced := debounce.NewWithTimeout(100*time.Millisecond, 0)

		for range 5 {
			debounced(func() { calls.Add(1) })
			time.Sleep(50 * time.Millisecond)
		}

		// Continuous calls must keep postponing the function when the timeout
		// is disabled.
		if got := calls.Load(); got != 0 {
			t.Fatalf("expected no run while calls continue, got %d", got)
		}

		time.Sleep(110 * time.Millisecond)
		if got := calls.Load(); got != 1 {
			t.Fatalf("expected one run after calls stop, got %d", got)
		}
	})
}
