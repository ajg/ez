// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"runtime"
	"testing"
)

// A Unit is a specification that can be used for testing and/or benchmarking.
type Unit struct {
	fn interface{}
	rs []runner
	T  *testing.T
	B  *testing.B
	tr bool
	br bool
}

// A half is an incomplete case.
type half struct {
	in in
	u  *Unit
}

// A runner is one of many sequential components of a Unit.
type runner interface {
	runTest(int, *testing.T)
	runBenchmark(int, *testing.B)
}

// New returns a blank Unit.
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

// Test returns a Unit for testing fn using t.
func Test(fn interface{}, t *testing.T) *Unit {
	return New().setT(t).Func(fn)
}

// Benchmark returns a Unit for benchmarking fn using b.
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

// Func sets fn as the Unit's current function; it can be called more than once.
func (u *Unit) Func(fn interface{}) *Unit {
	u.fn = fn
	return u
}

// Thru applies fn to the Unit, which is useful to combine or repeat common Unit components.
func (u *Unit) Thru(fn func(*Unit)) *Unit {
	fn(u)
	return u
}

// Step adds fn to the Unit as a Step.
func (u *Unit) Step(fn func()) *Unit {
	u.rs = append(u.rs, Step{fn})
	return u
}

// Case adds in & out (plus the current function) as a Case to the Unit.
func (u *Unit) Case(in in, out out) *Unit {
	return u.addCase(in, out)
}

// Cases adds every in/out pair in the CaseMap (plus the current function) as a Case to the Unit.
func (u *Unit) Cases(cs CaseMap) *Unit {
	for in, out := range cs {
		u = u.addCase(*in, out)
	}
	return u
}

// In returns xs as inputs that can be used in a Case or CaseMap.
func In(xs ...interface{}) *in { in := newIn(xs); return &in }

// Out returns xs as outputs that can be used in a Case or CaseMap.
func Out(xs ...interface{}) out { return newOut(xs) }

// Panic returns a requirement to panic with any value, and can be used in a Case or CaseMap; it is equivalent to PanicWith(Any).
func Panic() out { return newPanic(Any) }

// PanicWith returns a requirement to panic with x, and can be used in a Case or CaseMap.
func PanicWith(x interface{}) out { return newPanic(x) }

// In begins a Case with xs as inputs.
func (u *Unit) In(xs ...interface{}) *half { return &half{newIn(xs), u} }

// Out completes a Case with xs as outputs, and adds it to the Unit.
func (h *half) Out(xs ...interface{}) *Unit { return h.u.addCase(h.in, newOut(xs)) }

// Panic completes a Case that must panic with any value, and adds it to the Unit; it is equivalent to PanicWith(Any).
func (h *half) Panic() *Unit { return h.u.addCase(h.in, newPanic(Any)) }

// PanicWith completes a Case that must panic with x, and adds it to the Unit.
func (h *half) PanicWith(x interface{}) *Unit { return h.u.addCase(h.in, newPanic(x)) }

func (u *Unit) addCase(in in, out out) *Unit {
	u.rs = append(u.rs, newCase(u.fn, in, out))
	return u
}

// Run runs the Unit as a test and/or benchmark, depending on whether T and/or B are set.
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

// RunTest runs the Unit as a test using t.
func (u *Unit) RunTest(t *testing.T) {
	if u.tr {
		panic("test already ran")
	}
	u.tr = true

	for i, r := range u.rs {
		r.runTest(i, t)
	}
}

// RunBenchmark runs the Unit as a benchmark using b.
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
