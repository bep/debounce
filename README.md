# Go Debounce

[![Build Status](https://travis-ci.org/bep/debounce.svg)](https://travis-ci.org/bep/debounce)
[![GoDoc](https://godoc.org/github.com/bep/debounce?status.svg)](https://godoc.org/github.com/bep/debounce)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/debounce)](https://goreportcard.com/report/github.com/bep/debounce)
[![codecov](https://codecov.io/gh/bep/debounce/branch/master/graph/badge.svg)](https://codecov.io/gh/bep/debounce)
[![Release](https://img.shields.io/github/release/bep/debounce.svg?style=flat-square)](https://github.com/bep/debounce/releases/latest)

## Why?

This may seem like a fairly narrow library, so why not copy-and-paste it when needed? Sure -- but this is, however, slightly more usable than [left-pad](https://www.npmjs.com/package/left-pad), and as I move my client code into the [GopherJS](https://github.com/gopherjs/gopherjs) world, a [debounce](https://davidwalsh.name/javascript-debounce-function) function is a must-have.

This library works, but if you find any issue or a potential improvement, please create an issue or a pull request!

## Use

This package provides a debouncer func. The most typical use case would be the user 
typing a text into a form; the UI needs an update, but let's wait for a break.

`New` returns a debounced function and a channel that can be closed to signal a stop of the goroutine. The function will, as long as it continues to be invoked, not be triggered. The function will be called after it stops being called for the given duration. Note that a stop signal means a full stop of the debouncer; there is no concept of flushing future invocations. 

**Note:** The created debounced function can be invoked with different functions, if needed, the last one will win.

An example:

```go
func ExampleNew() {
	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced, finish, done := debounce.New(100 * time.Millisecond)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f)
		}

		time.Sleep(200 * time.Millisecond)
	}

	close(finish)

	<-done

	c := int(atomic.LoadUint64(&counter))

	fmt.Println("Counter is", c)
	// Output: Counter is 3
}
```

## Tests

To run the tests, you need to install `Leaktest`:

```bash
go get -u github.com/fortytw2/leaktest
```