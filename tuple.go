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

type tuple struct {
	xs []interface{}
	f  string
	l  int
	// TODO: p interface{} for panics
}

func newTuple(xs []interface{}) *tuple {
	_, f, l, ok := runtime.Caller(2) // newTuple + In/Out/Panic
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
	return &tuple{xs, f, l}
}

func (t tuple) Equal(u tuple) bool {
	return reflect.DeepEqual(t.xs, u.xs)
}

func (t tuple) values(f reflect.Value) (vs []reflect.Value) {
	ft := f.Type()

	for i, x := range t.xs {
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
