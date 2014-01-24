// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"runtime"
	"testing"
)

type Unit struct {
	fn interface{}
	rs []runner
	T  *testing.T
	B  *testing.B
	tr bool
	br bool
}

type half struct {
	in in
	u  *Unit
}

type runner interface {
	runTest(int, *testing.T)
	runBenchmark(int, *testing.B)
}

func New() *Unit {
	u := &Unit{}
	// FIXME: Sadly, finalizers are not guaranteed to run, so they're of little comfort.
	runtime.SetFinalizer(u, func(u *Unit) {
		if !u.tr && !u.br {
			panic("neither test nor benchmark ran")
		}
	})
	return u
}

func Test(fn interface{}, t *testing.T) *Unit {
	return New().setT(t).Func(fn)
}

func Benchmark(fn interface{}, b *testing.B) *Unit {
	return New().setB(b).Func(fn)
}

func (u *Unit) setT(t *testing.T) *Unit {
	u.T = t
	return u
}

func (u *Unit) setB(b *testing.B) *Unit {
	u.B = b
	return u
}

func (u *Unit) Func(fn interface{}) *Unit {
	u.fn = fn
	return u
}

func (u *Unit) Thru(fn func(*Unit)) *Unit {
	fn(u)
	return u
}

func (u *Unit) Step(fn func()) *Unit {
	u.rs = append(u.rs, Step{fn})
	return u
}

func (u *Unit) Case(in in, out out) *Unit {
	return u.addCase(in, out)
}

func (u *Unit) Cases(cs CaseMap) *Unit {
	for in, out := range cs {
		u = u.addCase(*in, out)
	}
	return u
}

func In(xs ...interface{}) in   { return newIn(xs) }
func Out(xs ...interface{}) out { return newOut(xs) }
func Panic(x interface{}) out   { return newPanic(x) }

func (u *Unit) In(xs ...interface{}) *half  { return &half{newIn(xs), u} }
func (h *half) Out(xs ...interface{}) *Unit { return h.u.addCase(h.in, newOut(xs)) }
func (h *half) Panic(x interface{}) *Unit   { return h.u.addCase(h.in, newPanic(x)) }

func (u *Unit) addCase(in in, out out) *Unit {
	u.rs = append(u.rs, newCase(u.fn, in, out))
	return u
}

func (u *Unit) Run() {
	if u.B == nil && u.T == nil {
		panic("T and B are both nil")
	}
	if u.T != nil {
		u.RunTest(u.T)
	}
	if u.B != nil {
		u.RunBenchmark(u.B)
	}
}

func (u *Unit) RunTest(t *testing.T) {
	if u.tr {
		panic("test already ran")
	}
	u.tr = true

	for i, r := range u.rs {
		r.runTest(i, t)
	}
}

func (u *Unit) RunBenchmark(b *testing.B) {
	if u.br {
		panic("benchmark already ran")
	}
	u.br = true

	for i := 0; i < b.N; i++ {
		for j, r := range u.rs {
			r.runBenchmark(j, b)
		}
	}
}
