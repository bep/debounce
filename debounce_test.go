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

	counter1 := 0

	f1 := func() {
		counter1++
	}

	counter2 := 0

	f2 := func() {
		counter2++
	}

	f3 := func() {
		counter2 += 2
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

	if counter1 != 3 {
		t.Error("Expected count 3, was", counter1)
	}

	if counter2 != 8 {
		t.Error("Expected count 8, was", counter2)
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

	if counter != 21 {
		t.Error("Expected count 21, was", counter)
	}
}

func TestDebounceCloseEarly(t *testing.T) {
	defer leaktest.Check(t)()

	counter := 0

	f := func() {
		counter++
	}

	debounced, finish := debounce.New(100 * time.Millisecond)

	debounced(f)

	close(finish)

	<-time.After(200 * time.Millisecond)

	if counter != 0 {
		t.Error("Expected count 0, was", counter)
	}

}

func BenchmarkDebounce(b *testing.B) {
	counter := 0

	f := func() {
		counter++
	}

	debounced, finish := debounce.New(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debounced(f)
	}
	close(finish)
	if counter != 0 {
		b.Fatal("Expected count 0, was", counter)
	}
}

func ExampleNew() {
	counter := 0

	f := func() {
		counter++
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

	fmt.Println("Counter is", counter)
	// Output: Counter is 3
}
