EZ
==

A Go Package for Easy, Powerful Tests and Benchmarks, written by [Alvaro J. Genial](http://alva.ro).

[![Build Status](https://travis-ci.org/ajg/ez.png?branch=master)](https://travis-ci.org/ajg/ez)
[![GoDoc](https://godoc.org/github.com/ajg/ez?status.png)](https://godoc.org/github.com/ajg/ez)

Synopsis
--------

The purpose of this library is to facilitate the testing and benchmarking of Go code. It accomplishes this goal by...

 - Reducing the amount of boilerplate needed to a mininum.
 - Eliminating most of the manual, error-prone parts of testing.
 - Allowing testing code to be automatically reused for benchmarking.
 - Providing a series of useful, contextual information automatically.
 - Focusing on what is being tested rather than how it is done.

...while remaining compatible with the standard `testing` package (used by the `go test` tool) as well as preserving its lightweight approach.

Status
------

This library is still very experimental and thus may change greatly in both interface and implementation. Please refrain from using it a production setting unless breaking changes can be tolerated.

Dependencies
------------

The only requirement is [Go 1.0](http://golang.org/doc/go1) or later.

License
-------

This library is distributed under a BSD-style [LICENSE](./LICENSE).
