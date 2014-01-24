// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"reflect"
	"runtime"
	"testing"
)

type Unit struct {
	f  *reflect.Value
	rs []runner
	T  *testing.T
	B  *testing.B
	tr bool
	br bool
}

type half struct {
	in tuple
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
	return NewUnit().setT(t).Of(fn)
}

func Benchmark(fn interface{}, b *testing.B) *Unit {
	return NewUnit().setB(b).Of(fn)
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
	f := reflect.ValueOf(fn)
	if !f.IsValid() || f.Kind() != reflect.Func {
		panic("not a valid function")
	}
	u.f = &f
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

func (u *Unit) Case(in, out *tuple) *Unit {
	return u.addCase(*in, out)
}

func (u *Unit) Cases(cs CaseMap) *Unit {
	for in, out := range cs {
		u = u.addCase(*in, out)
	}
	return u
}

func In(xs ...interface{}) *tuple  { return newTuple(xs) }
func Out(xs ...interface{}) *tuple { return newTuple(xs) }
func Panic() *tuple                { return nil }

func (u *Unit) In(xs ...interface{}) *half  { return &half{*newTuple(xs), u} }
func (h *half) Out(xs ...interface{}) *Unit { return h.u.addCase(h.in, newTuple(xs)) }
func (h *half) Panic() *Unit                { return h.u.addCase(h.in, nil) }

func (u *Unit) addCase(in tuple, out *tuple) *Unit {
	u.rs = append(u.rs, u.newCase(in, out))
	return u
}

func (u *Unit) newCase(in tuple, out *tuple) Case {
	if u.f == nil {
		panic("test has no function")
	}
	return Case{*u.f, in, out}
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
