package debounce

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkNew measures the cost of creating a new debounced function
func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(100 * time.Millisecond)
	}
}

// BenchmarkNewWithOptions measures the cost of creating with options
func BenchmarkNewWithOptions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(100*time.Millisecond, WithMaxCalls(5), WithMaxWait(1*time.Second))
	}
}

// BenchmarkSingleCall measures the cost of a single debounced call
func BenchmarkSingleCall(b *testing.B) {
	debounced := New(100 * time.Millisecond)
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		debounced(fn)
	}
}

// BenchmarkMultipleCalls measures the cost of multiple rapid calls
func BenchmarkMultipleCalls(b *testing.B) {
	debounced := New(100 * time.Millisecond)
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Make 10 calls per iteration to simulate rapid calling
		for j := 0; j < 10; j++ {
			debounced(fn)
		}
	}
}

// BenchmarkCallLimitTrigger measures performance when call limit is reached
func BenchmarkCallLimitTrigger(b *testing.B) {
	debounced := New(1*time.Second, WithMaxCalls(5))
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Make exactly 5 calls to trigger immediate execution
		for j := 0; j < 5; j++ {
			debounced(fn)
		}
	}
}

// BenchmarkTimeLimitTrigger measures performance when time limit is reached
func BenchmarkTimeLimitTrigger(b *testing.B) {
	var wg sync.WaitGroup

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		debounced := New(1*time.Second, WithMaxWait(1*time.Millisecond))

		wg.Add(1)
		fn := func() {
			wg.Done()
		}

		debounced(fn)
		wg.Wait()
	}
}

// BenchmarkConcurrentCalls measures performance under concurrent access
func BenchmarkConcurrentCalls(b *testing.B) {
	debounced := New(100 * time.Millisecond)
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			debounced(fn)
		}
	})
}

// BenchmarkConcurrentCallsWithLimits measures concurrent performance with limits
func BenchmarkConcurrentCallsWithLimits(b *testing.B) {
	debounced := New(100*time.Millisecond, WithMaxCalls(10), WithMaxWait(50*time.Millisecond))
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			debounced(fn)
		}
	})
}

// BenchmarkMemoryUsage measures memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

	debounced := New(100 * time.Millisecond)
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		debounced(fn)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
}

// BenchmarkTimerCreation measures the cost of timer creation and cancellation
func BenchmarkTimerCreation(b *testing.B) {
	debounced := New(1 * time.Millisecond)
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Each call creates and potentially cancels a timer
		debounced(fn)
		debounced(fn) // This will cancel the previous timer
	}
}

// BenchmarkDifferentFunctions measures cost when different functions are passed
func BenchmarkDifferentFunctions(b *testing.B) {
	debounced := New(100 * time.Millisecond)

	functions := []func(){
		func() { _ = 1 + 1 },
		func() { _ = "hello" + "world" },
		func() { _ = make([]int, 10) },
		func() { _ = map[string]int{"test": 1} },
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		debounced(functions[i%len(functions)])
	}
}

// BenchmarkScaling tests performance with different numbers of goroutines
func BenchmarkScaling(b *testing.B) {
	goroutineCounts := []int{1, 2, 4, 8, 16, 32, 64, 128}

	for _, numGoroutines := range goroutineCounts {
		b.Run(fmt.Sprintf("goroutines-%d", numGoroutines), func(b *testing.B) {
			debounced := New(100 * time.Millisecond)
			fn := func() {}

			b.ResetTimer()
			b.ReportAllocs()

			var wg sync.WaitGroup
			callsPerGoroutine := b.N / numGoroutines

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < callsPerGoroutine; j++ {
						debounced(fn)
					}
				}()
			}
			wg.Wait()
		})
	}
}

// BenchmarkWithVariousDurations tests performance with different debounce durations
func BenchmarkWithVariousDurations(b *testing.B) {
	durations := []time.Duration{
		1 * time.Nanosecond,
		1 * time.Microsecond,
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		1 * time.Second,
	}

	for _, duration := range durations {
		b.Run(fmt.Sprintf("duration-%v", duration), func(b *testing.B) {
			debounced := New(duration)
			fn := func() {}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				debounced(fn)
			}
		})
	}
}

// BenchmarkCallLimitVariations tests performance with different call limits
func BenchmarkCallLimitVariations(b *testing.B) {
	limits := []int{1, 2, 5, 10, 50, 100, -1}

	for _, limit := range limits {
		name := fmt.Sprintf("limit-%d", limit)
		if limit == -1 {
			name = "limit-none"
		}

		b.Run(name, func(b *testing.B) {
			debounced := New(100*time.Millisecond, WithMaxCalls(limit))
			fn := func() {}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				debounced(fn)
			}
		})
	}
}

// BenchmarkReset measures the cost of counter and timer reset after execution
func BenchmarkReset(b *testing.B) {
	debounced := New(1*time.Nanosecond, WithMaxCalls(1)) // Immediate execution
	var wg sync.WaitGroup

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		debounced(func() {
			wg.Done()
		})
		wg.Wait() // Wait for execution and reset
	}
}

// BenchmarkHighFrequency simulates high-frequency scenarios
func BenchmarkHighFrequency(b *testing.B) {
	debounced := New(1*time.Millisecond, WithMaxCalls(1000))
	fn := func() {}

	b.ResetTimer()
	b.ReportAllocs()

	// Simulate burst of calls
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ { // 100 rapid calls per iteration
			debounced(fn)
		}
	}
}

// BenchmarkComparison compares debounced vs direct function calls
func BenchmarkComparison(b *testing.B) {
	fn := func() { _ = 1 + 1 }

	b.Run("direct-call", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fn()
		}
	})

	b.Run("debounced-call", func(b *testing.B) {
		debounced := New(100 * time.Millisecond)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			debounced(fn)
		}
	})
}
