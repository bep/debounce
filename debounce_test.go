package debounce_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bep/debounce"
	"github.com/fortytw2/leaktest"
)

func TestDebounce(t *testing.T) {
	defer leaktest.Check(t)()

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

	debounced, shutdown := debounce.New(100 * time.Millisecond)

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

	close(shutdown)

	<-time.After(200 * time.Millisecond)

	c1 := int(atomic.LoadUint64(&counter1))
	c2 := int(atomic.LoadUint64(&counter2))
	if c1 != 3 {
		t.Error("Expected count 3, was", c1)
	}
	if c2 != 8 {
		t.Error("Expected count 8, was", c2)
	}
}

func TestDebounceInParallel(t *testing.T) {
	defer leaktest.Check(t)()

	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced, shutdown := debounce.New(100 * time.Millisecond)

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debouncedInner, shutdown := debounce.New(100 * time.Millisecond)
			for j := 0; j < 10; j++ {
				debouncedInner(f)
				debounced(f)
			}
			time.Sleep(150 * time.Millisecond)
			close(shutdown)
		}()
	}
	wg.Wait()

	close(shutdown)

	<-time.After(200 * time.Millisecond)

	c := int(atomic.LoadUint64(&counter))
	if c != 21 {
		t.Error("Expected count 21, was", c)
	}
}

func TestDebounceCloseEarly(t *testing.T) {
	defer leaktest.Check(t)()

	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced, finish := debounce.New(100 * time.Millisecond)

	debounced(f)

	close(finish)

	<-time.After(200 * time.Millisecond)

	c := int(atomic.LoadUint64(&counter))
	if c != 0 {
		t.Error("Expected count 0, was", c)
	}

}

func BenchmarkDebounce(b *testing.B) {
	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced, finish := debounce.New(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debounced(f)
	}
	close(finish)
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

	debounced, finish := debounce.New(100 * time.Millisecond)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f)
		}

		time.Sleep(200 * time.Millisecond)
	}

	close(finish)

	<-time.After(200 * time.Millisecond)

	c := int(atomic.LoadUint64(&counter))

	fmt.Println("Counter is", c)
	// Output: Counter is 3
}
