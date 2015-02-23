// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type tuple struct {
	xs []interface{}
}

type in struct {
	t tuple
	f string
	l int
}

type out struct {
	t tuple
	f string
	l int
	p interface{}
}

type (
	ANY   struct{}
	NIL   struct{}
	ZERO  struct{}
	ERROR struct{}
)

// Any is a placeholder that can be used with Out and Panic, and it means any value is acceptable.
var Any = ANY{}

// Nil is a placeholder that can be used with Out and Panic, and it means only nil is acceptable.
var Nil = NIL{}

// Zero is a placeholder that can be used with Out and Panic, and it means only the zero value is acceptable.
var Zero = ZERO{}

// Error is a placeholder that can be used with Out and Panic, and it means any error is acceptable.
var Error = ERROR{}

func (t tuple) String() string {
	s := "("
	for i, x := range t.xs {
		if i != 0 {
			s += ", "
		}
		s += fmt.Sprintf("%#+v", x)
	}
	return s + ")"
}

func (t tuple) GoString() string {
	return t.String()
}

func newIn(xs []interface{}) in {
	f, l := source()
	return in{tuple{xs}, f, l}
}

func newOut(xs []interface{}) out {
	f, l := source()
	return out{tuple{xs}, f, l, nil}
}

func newPanic(x interface{}) out {
	f, l := source()
	return out{tuple{}, f, l, x}
}

func (t tuple) equal(u tuple) bool {
	switch {
	case t.xs == nil && u.xs != nil ||
		t.xs != nil && u.xs == nil ||
		len(t.xs) != len(u.xs):
		return false
	}
	for i, x := range t.xs {
		if y := u.xs[i]; !areEqual(x, y) {
			return false
		}
	}
	return true
}

// TODO: Use mirror for this.
func isAny(x interface{}) bool { return x == Any }
func isNil(x interface{}) bool { return x == nil || x == Nil }
func isZero(x interface{}) bool {
	if x == Zero {
		return true
	}
	return false // TODO: Use reflection.
}
func isError(x interface{}) bool { _, ok := x.(error); return ok || x == Error }

func areEqual(x, y interface{}) (b bool) {
	/*	b = areEqual1(x, y)
			log.Println("areEqual", fmt.Sprintf("%#+v", x), "==", fmt.Sprintf("%#+v", y), "=>", b)
			return b
		}

		func areEqual1(x, y interface{}) bool {*/
	switch {
	case isAny(x) || isAny(y):
		return true
	case isNil(x) && isNil(y):
		return true
	case isZero(x) && isZero(y):
		return true
	case isError(x) && isError(y):
		return true
	}
	return reflect.DeepEqual(x, y)
}

func (in in) values(f reflect.Value) (vs []reflect.Value) {
	ft := f.Type()

	for i, x := range in.t.xs {
		var t reflect.Type
		if n := ft.NumIn(); i < n {
			t = ft.In(i)
		} else if ft.IsVariadic() {
			t = ft.In(n - 1).Elem()
		} else {
			panic("too many input values")
		}

		if v, ok := validValueOrZero(reflect.ValueOf(x), t); !ok {
			panic("invalid input value")
		} else {
			vs = append(vs, v)
		}
	}
	return vs
}

func validValueOrZero(v reflect.Value, t reflect.Type) (reflect.Value, bool) {
	if v.IsValid() {
		return v, true
	}
	switch k := t.Kind(); k {
	case reflect.Ptr:
		// This can happen when passing an untyped nil, so we'll do the caller a favor and type it for them.
		return reflect.Zero(t), true
	case reflect.Interface:
		// This can happens when passing in a nil value of interface type, because it's lost
		// at runtime. See https://groups.google.com/forum/#!topic/golang-nuts/qgJy_H2GysY
		// We'll also be nice and provide them with a proper nil value of that type.
		return reflect.Zero(t), true
	}
	return reflect.Value{}, false
}

func source() (string, int) {
	_, f, l, ok := runtime.Caller(3) // source + newIn/newOut/newPanic + In/Out/Panic
	if ok {
		// Truncate file name at last file name separator.
		if i := strings.LastIndex(f, "/"); i >= 0 {
			f = f[i+1:]
		} else if i = strings.LastIndex(f, "\\"); i >= 0 {
			f = f[i+1:]
		}
	} else {
		f = "???"
		l = 1
	}
	return f, l
}
