package ez

import (
	"reflect"
	"testing"
)

func TestIn(t *testing.T) {
	args := []interface{}{true, 42, "foo"}
	cin := in{tuple{args}, "unit_test.go", 11}
	if in := In(args...); !reflect.DeepEqual(cin, in) {
		t.Errorf("In(%v)\nwant (%v)\nhave (%v)", args, cin, in)
	}
}

func TestOut(t *testing.T) {
	args := []interface{}{true, 42, "foo"}
	cout := out{tuple{args}, "unit_test.go", 19, nil}
	if out := Out(args...); !reflect.DeepEqual(cout, out) {
		t.Errorf("Out(%v)\nwant (%v)\nhave (%v)", args, cout, out)
	}
}
