// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ez

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

type Case struct {
	f   reflect.Value
	in  tuple
	out *tuple // nil means panic
}

type CaseMap map[*tuple]*tuple

type Step struct {
	fn func()
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
	f := c.f
	n := runtime.FuncForPC(f.Pointer()).Name()
	defer func() {
		e := recover()
		if c.out == nil || e == nil {
			return
		}
		t.Errorf("case #%d %s - %s%v\n%s\n%s",
			i,
			colorf(black, white, " %s:%d ", c.in.f, c.in.l),
			n,
			c.in,
			colorf(green, black, "want %#+v", *c.out),
			colorf(red, black, "have panic [%s]\n%s", e, string(debug.Stack())),
		)
	}()
	if out := apply(f, c.in.values(f)); c.out == nil {
		t.Errorf("case #%d %s - %s%v\n%s\n%s",
			i,
			colorf(black, white, " %s:%d ", c.in.f, c.in.l),
			n,
			c.in,
			colorf(green, black, "want panic"), // TODO: Allow specifying the panic value or at least string.
			colorf(red, black, "have %#+v", out),
		)
	} else if !c.out.Equal(out) {
		t.Errorf("case #%d %s - %s%v\n%s\n%s\ndiff %s",
			i,
			colorf(black, white, " %s:%d ", c.in.f, c.in.l),
			n,
			c.in,
			colorf(green, black, "want %#+v", *c.out),
			colorf(red, black, "have %#+v", out),
			Diff(fmt.Sprintf("%#+v", out), fmt.Sprintf("%#+v", *c.out)),
		)
	}
}

func (c Case) runBenchmark(i int, b *testing.B) {
	b.StopTimer()
	args := c.in.values(c.f)
	b.StartTimer()
	c.f.Call(args)
}

func apply(f reflect.Value, args []reflect.Value) tuple {
	var ys []interface{}
	for _, v := range f.Call(args) {
		ys = append(ys, v.Interface())
	}
	return tuple{ys, "", 0}
}

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

var Diff = func(a, b string) (s string) {
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
	bs, err := exec.Command("git", "diff", "--color-words", "--no-index", af.Name(), bf.Name()).Output()
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
