// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"errors"
	"reflect"
	"testing"
)

func TestIn_Abstract(t *testing.T) {
	PathStyle = Abstract
	args := []interface{}{true, 42, "foo"}
	cin := in{tuple{args}, "github.com/ajg/ez/unit_test.go", 17}
	if in := In(args...); !reflect.DeepEqual(cin, *in) {
		t.Errorf("In(%v)\nwant (%v)\nhave (%v)", args, cin, *in)
	}
}

func TestOut_Abstract(t *testing.T) {
	PathStyle = Abstract
	args := []interface{}{true, 42, "foo"}
	cout := out{tuple{args}, "github.com/ajg/ez/unit_test.go", 26, nil}
	if out := Out(args...); !reflect.DeepEqual(cout, out) {
		t.Errorf("Out(%v)\nwant (%v)\nhave (%v)", args, cout, out)
	}
}

func TestPanic_Abstract(t *testing.T) {
	PathStyle = Abstract
	cout := out{tuple{}, "github.com/ajg/ez/unit_test.go", 34, Any}
	if out := Panic(); !reflect.DeepEqual(cout, out) {
		t.Errorf("Panic()\nwant (%v)\nhave (%v)", cout, out)
	}
}

func TestPanicWith_Abstract(t *testing.T) {
	PathStyle = Abstract
	p := errors.New("bar")
	cout := out{tuple{}, "github.com/ajg/ez/unit_test.go", 43, p}
	if out := PanicWith(p); !reflect.DeepEqual(cout, out) {
		t.Errorf("PanicWith(%v)\nwant (%v)\nhave (%v)", p, cout, out)
	}
}

func TestIn_Truncate(t *testing.T) {
	PathStyle = Truncate
	args := []interface{}{true, 42, "foo"}
	cin := in{tuple{args}, "unit_test.go", 52}
	if in := In(args...); !reflect.DeepEqual(cin, *in) {
		t.Errorf("In(%v)\nwant (%v)\nhave (%v)", args, cin, *in)
	}
}

func TestOut_Truncate(t *testing.T) {
	PathStyle = Truncate
	args := []interface{}{true, 42, "foo"}
	cout := out{tuple{args}, "unit_test.go", 61, nil}
	if out := Out(args...); !reflect.DeepEqual(cout, out) {
		t.Errorf("Out(%v)\nwant (%v)\nhave (%v)", args, cout, out)
	}
}

func TestPanic_Truncate(t *testing.T) {
	PathStyle = Truncate
	cout := out{tuple{}, "unit_test.go", 69, Any}
	if out := Panic(); !reflect.DeepEqual(cout, out) {
		t.Errorf("Panic()\nwant (%v)\nhave (%v)", cout, out)
	}
}

func TestPanicWith_Truncate(t *testing.T) {
	PathStyle = Truncate
	p := errors.New("bar")
	cout := out{tuple{}, "unit_test.go", 78, p}
	if out := PanicWith(p); !reflect.DeepEqual(cout, out) {
		t.Errorf("PanicWith(%v)\nwant (%v)\nhave (%v)", p, cout, out)
	}
}
