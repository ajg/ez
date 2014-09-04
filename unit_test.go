// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"errors"
	"reflect"
	"testing"
)

func TestIn(t *testing.T) {
	args := []interface{}{true, 42, "foo"}
	cin := in{tuple{args}, "unit_test.go", 16}
	if in := In(args...); !reflect.DeepEqual(cin, *in) {
		t.Errorf("In(%v)\nwant (%v)\nhave (%v)", args, cin, in)
	}
}

func TestOut(t *testing.T) {
	args := []interface{}{true, 42, "foo"}
	cout := out{tuple{args}, "unit_test.go", 24, nil}
	if out := Out(args...); !reflect.DeepEqual(cout, out) {
		t.Errorf("Out(%v)\nwant (%v)\nhave (%v)", args, cout, out)
	}
}

func TestPanic(t *testing.T) {
	p := errors.New("bar")
	cout := out{tuple{}, "unit_test.go", 32, p}
	if out := Panic(p); !reflect.DeepEqual(cout, out) {
		t.Errorf("Panic(%v)\nwant (%v)\nhave (%v)", p, cout, out)
	}
}
