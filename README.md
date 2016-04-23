# Go Debounce

[![Build Status](https://travis-ci.org/bep/debounce.svg)](https://travis-ci.org/bep/debounce)
[![GoDoc](https://godoc.org/github.com/bep/debounce?status.svg)](https://godoc.org/github.com/bep/debounce)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/debounce)](https://goreportcard.com/report/github.com/bep/debounce)
[![Coverage](http://gocover.io/_badge/github.com/bep/debounce)](http://gocover.io/github.com/bep/debounce)

## Why?

This may seem like a fairly narrow library, so why not copy-and-paste it when needed? Sure -- but this is, however, slightly more usable than [left-pad](https://www.npmjs.com/package/left-pad), and as I move my client code into the [GopherJS](https://github.com/gopherjs/gopherjs) world, a [debounce](https://davidwalsh.name/javascript-debounce-function) function is a must-have.

This library works, but if you find any issue or a potential improvement, please create an issue or a pull request!

## Use

This package provides a debouncer func. The most typical use case would be the user 
typing a text into a form; the UI needs an update, but let's wait for a break.

`New` returns a debounced function and a channel that can be closed to signal a stop of the goroutine. The function will, as long as it continues to be invoked, not be triggered. The function will be called after it stops being called for the given duration. Note that a stop signal means a full stop of the debouncer; there is no concept of flushing future invocations. 

An example:

```go
	counter := 0
	
	f := func() {
		counter++
	}

	debounced, finish := debounce.New(2*time.Second)

	for i := 0; i < 10; i++ {
		debounced(f)
	}

	time.Sleep(3 * time.Second)
    
    close(finish)
	
	<-time.After(200 * time.Millisecond)

	if counter != 1 {
		panic("Count mismatch")
	}
```

## Tests

To run the tests, you need to install `Leaktest`:

```bash
go get -u github.com/fortytw2/leaktest
```