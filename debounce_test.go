package debounce

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	debounced := New(100 * time.Millisecond)
	if debounced == nil {
		t.Fatal("New() returned nil")
	}
}

func TestBasicDebounce(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(50 * time.Millisecond)

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// Call multiple times quickly
	debounced(fn)
	debounced(fn)
	debounced(fn)

	// Should not be called yet
	mu.Lock()
	if called != 0 {
		t.Errorf("Expected 0 calls, got %d", called)
	}
	mu.Unlock()

	// Wait for debounce period
	time.Sleep(100 * time.Millisecond)

	// Should be called once
	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestDebounceCancellation(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(100 * time.Millisecond)

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// First call
	debounced(fn)

	// Wait half the debounce time
	time.Sleep(50 * time.Millisecond)

	// Call again - should cancel the first timer
	debounced(fn)

	// Wait for the original debounce time
	time.Sleep(60 * time.Millisecond)

	// Should not be called yet
	mu.Lock()
	if called != 0 {
		t.Errorf("Expected 0 calls, got %d", called)
	}
	mu.Unlock()

	// Wait for the second debounce time
	time.Sleep(50 * time.Millisecond)

	// Should be called once
	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestWithMaxCalls(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(100*time.Millisecond, WithMaxCalls(3))

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// Call exactly the limit number of times
	debounced(fn)
	debounced(fn)
	debounced(fn)

	// Should be called immediately when limit is reached
	time.Sleep(10 * time.Millisecond) // Small delay to allow execution

	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestWithMaxCallsNoLimit(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(50*time.Millisecond, WithMaxCalls(-1))

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// Call many times
	for i := 0; i < 10; i++ {
		debounced(fn)
	}

	// Should not be called immediately
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if called != 0 {
		t.Errorf("Expected 0 calls, got %d", called)
	}
	mu.Unlock()

	// Wait for debounce
	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestWithMaxWait(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(200*time.Millisecond, WithMaxWait(100*time.Millisecond))

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// First call
	debounced(fn)

	// Keep calling before debounce timeout but within max wait
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)
		debounced(fn)
	}

	// Should be called due to max wait limit
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestCombinedLimits(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(200*time.Millisecond, WithMaxCalls(2), WithMaxWait(100*time.Millisecond))

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// Call twice to hit call limit
	debounced(fn)
	debounced(fn)

	// Should be called immediately due to call limit
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestLastFunctionWins(t *testing.T) {
	var result string
	var mu sync.Mutex

	debounced := New(50 * time.Millisecond)

	fn1 := func() {
		mu.Lock()
		result = "first"
		mu.Unlock()
	}

	fn2 := func() {
		mu.Lock()
		result = "second"
		mu.Unlock()
	}

	debounced(fn1)
	debounced(fn2)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if result != "second" {
		t.Errorf("Expected 'second', got '%s'", result)
	}
	mu.Unlock()
}

func TestConcurrentCalls(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(50 * time.Millisecond)

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	var wg sync.WaitGroup

	// Launch multiple goroutines calling the debounced function
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debounced(fn)
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestResetBehavior(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(100*time.Millisecond, WithMaxCalls(3))

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	// First batch - hit call limit
	debounced(fn)
	debounced(fn)
	debounced(fn)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	firstCalled := called
	mu.Unlock()

	if firstCalled != 1 {
		t.Errorf("Expected 1 call after first batch, got %d", firstCalled)
	}

	// Second batch - should reset and work again
	debounced(fn)
	debounced(fn)
	debounced(fn)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if called != 2 {
		t.Errorf("Expected 2 calls total, got %d", called)
	}
	mu.Unlock()
}

func TestZeroDuration(t *testing.T) {
	var called int
	var mu sync.Mutex

	debounced := New(0)

	fn := func() {
		mu.Lock()
		called++
		mu.Unlock()
	}

	debounced(fn)

	// With zero duration, should be called almost immediately
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if called != 1 {
		t.Errorf("Expected 1 call, got %d", called)
	}
	mu.Unlock()
}

func TestCallLimitReachedFunction(t *testing.T) {
	d := &debouncer{
		callsLimit: 3,
		calls:      3,
	}

	if !d.callLimitReached() {
		t.Error("Expected callLimitReached to return true when calls >= callsLimit")
	}

	d.calls = 2
	if d.callLimitReached() {
		t.Error("Expected callLimitReached to return false when calls < callsLimit")
	}

	d.callsLimit = -1
	d.calls = 100
	if d.callLimitReached() {
		t.Error("Expected callLimitReached to return false when callsLimit is -1 (no limit)")
	}
}

func TestTimeLimitReachedFunction(t *testing.T) {
	d := &debouncer{
		startWait: time.Now().Add(-200 * time.Millisecond),
		waitLimit: 100 * time.Millisecond,
	}

	if !d.timeLimitReached() {
		t.Error("Expected timeLimitReached to return true when time since startWait >= waitLimit")
	}

	d.startWait = time.Now()
	if d.timeLimitReached() {
		t.Error("Expected timeLimitReached to return false when time since startWait < waitLimit")
	}
}
