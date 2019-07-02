// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

// A Case represents a function and its required input and output values.
type Case struct {
	fn  interface{}
	in  in
	out out
}

// A CaseMap is a mapping from inputs to outputs.
type CaseMap map[*in]out

// A Step holds arbitrary code to execute during a specific part of a test.
type Step struct {
	fn func()
}

// Diff is a function that should produce a string representing the difference between two strings.
// By default it attempts to use the git command to produce an inline diff, which can be colorized.
var Diff = gitDiff

func newCase(fn interface{}, in in, out out) Case {
	if fn == nil {
		panic("function is nil")
	}
	return Case{fn, in, out}
}

func (s Step) runTest(int, *testing.T) {
	s.fn()
}

func (s Step) runBenchmark(_ int, b *testing.B) {
	b.StopTimer()
	s.fn()
	b.StartTimer()
}

func (c Case) runTest(i int, t *testing.T) {
	// TODO: Color i, n & c.in with default colors, so they can eventually be customized too.
	f := reflect.ValueOf(c.fn)
	if !f.IsValid() || f.Kind() != reflect.Func {
		panic("invalid function")
	}
	fn := runtime.FuncForPC(f.Pointer())
	d, n := splitName(fn.Name())
	n = d + "/" + colorf(black, white, "%s", n)
	s := colorf(white, black, " %s:%d ", c.in.f, c.in.l)

	defer func() {
		p := recover()
		switch {
		case p == nil || Any == c.out.p:
			return
		case c.out.p != nil:
			if reflect.DeepEqual(c.out.p, p) {
				return
			}
			// t.Errorf("step #%d %s\n - %s%v\n%s\n%s\ndiff %s", i,
			t.Errorf("%s\n - %s%v\n%s\n%s\ndiff %s",
				s,
				n,
				c.in.t,
				colorf(green, black, "want panic (%#+v)", c.out.p),
				colorf(red, black, "have panic (%#+v)\n%s", p, string(debug.Stack())),
				Diff(fmt.Sprintf("%# v", pretty.Formatter(p)),
					fmt.Sprintf("%# v", pretty.Formatter(c.out.p))),
			)
		default:
			// t.Errorf("step #%d %s\n - %s%v\n%s\n%s", i,
			t.Errorf("%s\n - %s%v\n%s\n%s",
				s,
				n,
				c.in.t,
				colorf(green, black, "want %#+v", c.out.t),
				colorf(red, black, "have panic [%s]\n%s", p, string(debug.Stack())),
			)
		}
	}()
	if out, err := apply(f, c.in.values(f), c.out.p != nil); err != nil {
		// t.Errorf("step #%d %s\n - %s%v\n%s\n%s", i,
		t.Errorf("%s\n - %s%v\n%s\n%s",
			s,
			n,
			c.in.t,
			colorf(red, black, "error [%s]", err.Error()),
			"",
		)
	} else if c.out.p != nil {
		// t.Errorf("step #%d %s\n - %s%v\n%s\n%s", i,
		t.Errorf("%s\n - %s%v\n%s\n%s",
			s,
			n,
			c.in.t,
			colorf(green, black, "want panic [%s]", c.out.p),
			colorf(red, black, "have %#+v", out),
		)
	} else if !c.out.t.equal(out) {
		// t.Errorf("step #%d %s\n - %s%v\n%s\n%s\ndiff %s", i,
		t.Errorf("%s\n - %s%v\n%s\n%s\ndiff %s",
			s,
			n,
			c.in.t,
			colorf(green, black, "want %#+v", pretty.Formatter(c.out.t)),
			colorf(red, black, "have %#+v", pretty.Formatter(out)),
			Diff(fmt.Sprintf("%# v", pretty.Formatter(out)),
				fmt.Sprintf("%# v", pretty.Formatter(c.out.t))),
		)
	}
}

func (c Case) runBenchmark(i int, b *testing.B) {
	b.StopTimer()
	f := reflect.ValueOf(c.fn)
	if !f.IsValid() || f.Kind() != reflect.Func {
		panic("invalid function")
	}
	args := c.in.values(f)

	b.StartTimer()
	f.Call(args)
}

func apply(f reflect.Value, args []reflect.Value, panicExpected bool) (_ tuple, err error) {
	if !panicExpected {
		defer func() {
			if e := recover(); e != nil {
				s := fmt.Sprint(e)
				err = errors.New(s)
				if !strings.HasSuffix(s, "reflect:") {
					log.Println("PANIC:")
					debug.PrintStack()
					log.Println("---")
				}
			}
		}()
	}

	var ys []interface{}
	for _, v := range f.Call(args) {
		ys = append(ys, v.Interface())
	}
	return tuple{ys}, err
}

func splitName(n string) (string, string) {
	// e.g. "example.org/user/package/foo.(Bar).Qux-fm" ~> "foo.Bar.Qux"
	d := path.Dir(n)
	n = path.Base(n)
	n = strings.TrimSuffix(n, "-fm")
	n = strings.Replace(n, "(", "", -1)
	n = strings.Replace(n, ")", "", -1)
	return d, n
}

// Colorize determines whether to attempt to use terminal colors.
var Colorize = true

const (
	white     = 15
	black     = 232
	gray      = 59 // 7
	green     = 40
	purple    = 60
	cyan      = 80
	orange    = 214
	yellow    = 226
	red       = 160
	brightRed = 196
)

func colorf(fg, bg uint16, format string, xs ...interface{}) string {
	s := fmt.Sprintf(format, xs...)
	if !Colorize {
		return s
	}
	code := func(a, b, c uint16) string { return fmt.Sprintf("%d;%d;%d", a, b, c) }
	return fmt.Sprintf("\033[%s;%sm%s\033[0m", code(38, 5, fg), code(48, 5, bg), s)
}

func gitDiff(a, b string) (s string) {
	defer func() {
		if e := recover(); e != nil {
			s = "<unavailable: please install git>" + "\n" + fmt.Sprint(e) + "\n" + string(debug.Stack())
		}
	}()

	dir := os.TempDir()
	af, err := ioutil.TempFile(dir, "A-")
	if err != nil {
		panic(err)
	}
	defer af.Close()
	bf, err := ioutil.TempFile(dir, "B-")
	if err != nil {
		panic(err)
	}
	defer bf.Close()
	if _, err = af.WriteString(a); err != nil {
		panic(err)
	}
	if _, err = bf.WriteString(b); err != nil {
		panic(err)
	}

	args := []string{"diff"}
	if Colorize {
		args = append(args, "--color-words")
	} else {
		args = append(args, "--word-diff", "--no-color")
	}
	args = append(args, "--no-index", af.Name(), bf.Name())

	bs, err := exec.Command("git", args...).Output()
	s = string(bs)
	if err != nil {
		// FIXME: Figure out how to make diff exit with 0 so that err is nil on
		//        success, otherwise we get "exit status 1".
		if len(s) == 0 {
			panic(err)
		}
	}

	if ss := strings.Split(s, "\n"); len(ss) >= 5 {
		// Skip the first five lines:
		// diff --git foo bar
		// index xyz
		// --- foo
		// +++ bar
		// @@
		return strings.Join(ss[5:], "\n")
	}
	return "<empty>"
}
