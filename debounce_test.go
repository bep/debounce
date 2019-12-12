package debounce_test

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bep/debounce"
	"github.com/stretchr/testify/assert"
)

func TestDebounce(t *testing.T) {
	t.Parallel()

	t.Run("immediate false", func(t *testing.T) {
		t.Parallel()

		testDebounce(t, false)
	})

	t.Run("immediate true", func(t *testing.T) {
		t.Parallel()

		testDebounce(t, true)
	})
}

func TestDebounceConcurrentAdd(t *testing.T) {
	t.Parallel()

	t.Run("immediate false", func(t *testing.T) {
		t.Parallel()

		testDebounceConcurrentAdd(t, false)
	})

	t.Run("immediate true", func(t *testing.T) {
		t.Parallel()

		testDebounceConcurrentAdd(t, true)
	})
}

func TestDebounceDelayed(t *testing.T) {
	t.Parallel()

	t.Run("immediate false", func(t *testing.T) {
		t.Parallel()

		testDebounceDelayed(t, false)
	})

	t.Run("immediate true", func(t *testing.T) {
		t.Parallel()

		testDebounceDelayed(t, true)
	})
}

func testDebounce(t *testing.T, immediate bool) {
	var (
		counter1 uint64
		counter2 uint64
	)

	f1 := func() {
		atomic.AddUint64(&counter1, 1)
	}

	f2 := func() {
		atomic.AddUint64(&counter2, 1)
	}

	f3 := func() {
		atomic.AddUint64(&counter2, 2)
	}

	debounced := debounce.New(100*time.Millisecond, immediate)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f1)
		}

		time.Sleep(200 * time.Millisecond)
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 10; j++ {
			debounced(f2)
		}
		for j := 0; j < 10; j++ {
			debounced(f3)
		}

		time.Sleep(200 * time.Millisecond)
	}

	c1 := int(atomic.LoadUint64(&counter1))
	c2 := int(atomic.LoadUint64(&counter2))

	if c1 != 3 {
		t.Error("Expected count 3, was", c1)
	}

	if !immediate {
		if c2 != 8 {
			t.Error("Expected count 8, was", c2)
		}
	} else {
		if c2 != 4 {
			t.Error("Expected count 4, was", c2)
		}
	}
}

func testDebounceConcurrentAdd(t *testing.T, immediate bool) {
	var wg sync.WaitGroup

	var flag uint64

	debounced := debounce.New(100*time.Millisecond, immediate)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debounced(func() {
				atomic.CompareAndSwapUint64(&flag, 0, 1)
			})
		}()
	}
	wg.Wait()

	time.Sleep(500 * time.Millisecond)
	c := int(atomic.LoadUint64(&flag))
	if c != 1 {
		t.Error("Flag not set")
	}
}

// Issue #1
func testDebounceDelayed(t *testing.T, immediate bool) {

	var (
		counter1 uint64
	)

	f1 := func() {
		atomic.AddUint64(&counter1, 1)
	}

	debounced := debounce.New(100*time.Millisecond, immediate)

	time.Sleep(110 * time.Millisecond)

	debounced(f1)

	time.Sleep(200 * time.Millisecond)

	c1 := int(atomic.LoadUint64(&counter1))
	if c1 != 1 {
		t.Error("Expected count 1, was", c1)
	}

}

func BenchmarkDebounce(b *testing.B) {
	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced := debounce.New(100*time.Millisecond, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debounced(f)
	}

	c := int(atomic.LoadUint64(&counter))
	if c != 0 {
		b.Fatal("Expected count 0, was", c)
	}
}

func ExampleNew() {
	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced := debounce.New(100*time.Millisecond, false)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f)
		}

		time.Sleep(200 * time.Millisecond)
	}

	c := int(atomic.LoadUint64(&counter))

	fmt.Println("Counter is", c)
	// Output: Counter is 3
}

func TestNewDebounce(t *testing.T) {
	t.Parallel()

	t.Run("immediate true", func(t *testing.T) {
		t.Parallel()

		d := 10 * time.Millisecond
		ticker := time.NewTicker(d / 10)
		timer := time.NewTimer(d)
		deb := debounce.New(d, true)
		buf := new(bytes.Buffer)
		f := func() {
			_, _ = fmt.Fprint(buf, "test")
		}
		deb(f)
		// f executed immediately
		assert.Equal(t, "test", buf.String())

	Loop:
		for {
			select {
			case <-ticker.C:
				deb(f)
			case <-timer.C:
				ticker.Stop()
				timer.Stop()
				break Loop
			}
		}

		assert.Equal(t, "test", buf.String())
	})

	t.Run("immediate false", func(t *testing.T) {
		d := 10 * time.Millisecond
		ticker := time.NewTicker(d / 10)
		timer := time.NewTimer(d)
		deb := debounce.New(d, false)
		buf := new(bytes.Buffer)
		f := func() {
			_, _ = fmt.Fprint(buf, "test")
		}
		deb(f)
		// f not executed now
		assert.NotEqual(t, "test", buf.String())

	Loop:
		for {
			select {
			case <-ticker.C:
				deb(f)
			case <-timer.C:
				ticker.Stop()
				timer.Stop()
				break Loop
			}
		}

		// slightly more time for ops
		time.Sleep(d + d/10)
		assert.Equal(t, "test", buf.String())
	})
}
