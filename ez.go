// Package ez provides an easy but powerful way to define unit tests and benchmarks that are compatible with package `testing`.
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

type Unit struct {
	f  *reflect.Value
	rs []runner
	T  *testing.T
	B  *testing.B
	tr bool
	br bool
}

type Case struct {
	f   reflect.Value
	in  tuple
	out *tuple // nil means panic
}

type Step struct {
	fn func()
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

type CaseMap map[*tuple]*tuple

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

type half struct {
	in tuple
	u  *Unit
}

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
		t.Errorf("\b \b \b case #%d %s - %s%v\n%s\n%s\ndiff %s",
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
