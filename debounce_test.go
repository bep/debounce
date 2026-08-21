// Copyright 2024 Bjørn Erik Pedersen
// SPDX-License-Identifier: MIT

package debounce_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bep/debounce"
)

func TestDebounce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			counter1 atomic.Uint64
			counter2 atomic.Uint64
		)

		f1 := func() {
			counter1.Add(1)
		}

		f2 := func() {
			counter2.Add(1)
		}

		f3 := func() {
			counter2.Add(2)
		}

		debounced := debounce.New(100 * time.Millisecond)

		for range 3 {
			for range 10 {
				debounced(f1)
			}

			time.Sleep(200 * time.Millisecond)
		}

		for range 4 {
			for range 10 {
				debounced(f2)
			}
			for range 10 {
				debounced(f3)
			}

			time.Sleep(200 * time.Millisecond)
		}

		c1 := int(counter1.Load())
		c2 := int(counter2.Load())
		if c1 != 3 {
			t.Error("Expected count 3, was", c1)
		}
		if c2 != 8 {
			t.Error("Expected count 8, was", c2)
		}
	})
}

func TestDebounceConcurrentAdd(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var wg sync.WaitGroup

		var flag atomic.Uint64

		debounced := debounce.New(100 * time.Millisecond)

		for range 10 {
			wg.Go(func() {
				debounced(func() {
					flag.CompareAndSwap(0, 1)
				})
			})
		}
		wg.Wait()

		time.Sleep(500 * time.Millisecond)
		c := int(flag.Load())
		if c != 1 {
			t.Error("Flag not set")
		}
	})
}

// Issue #1
func TestDebounceDelayed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			counter1 atomic.Uint64
		)

		f1 := func() {
			counter1.Add(1)
		}

		debounced := debounce.New(100 * time.Millisecond)

		time.Sleep(110 * time.Millisecond)

		debounced(f1)

		time.Sleep(200 * time.Millisecond)

		c1 := int(counter1.Load())
		if c1 != 1 {
			t.Error("Expected count 1, was", c1)
		}
	})
}

func BenchmarkDebounce(b *testing.B) {
	var counter atomic.Uint64

	f := func() {
		counter.Add(1)
	}

	debounced := debounce.New(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debounced(f)
	}

	c := int(counter.Load())
	if c != 0 {
		b.Fatal("Expected count 0, was", c)
	}
}

func ExampleNew() {
	var counter atomic.Uint64

	f := func() {
		counter.Add(1)
	}

	debounced := debounce.New(100 * time.Millisecond)

	for range 3 {
		for range 10 {
			debounced(f)
		}

		time.Sleep(200 * time.Millisecond)
	}

	c := int(counter.Load())

	fmt.Println("Counter is", c)
	// Output: Counter is 3
}

func TestDebounceCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var called atomic.Int32

		debounced, cancel := debounce.NewWithCancel(50 * time.Millisecond)

		// Schedule a call that would normally be executed.
		debounced(func() {
			called.Store(1)
		})

		// Cancel it before the timer is triggered.
		cancel()

		// Wait slightly longer than the debounce interval - if cancel did not work,
		//the function will execute and the test will fail.
		time.Sleep(70 * time.Millisecond)

		if called.Load() != 0 {
			t.Fatal("expected debounced function NOT to be called after cancel")
		}

		// Additionally, verify that calling cancel repeatedly is safe.
		cancel()
	})
}
