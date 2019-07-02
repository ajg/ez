package ez

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Note: this bit of inelegance is needed in situations such as running e.g. tests from a working
//       directory that isn't directly in GOPATH but that is ultimately reachable due to symlinks.

var anchors []string
var cwd string

func init() {
	cwd, _ = os.Getwd()

	var ps []string
	ps = append(ps, filepath.SplitList(build.Default.GOPATH)...)
	ps = append(ps, filepath.SplitList(runtime.GOROOT())...)
	for _, p := range ps {
		if p == "" {
			continue
		}
		src := filepath.Join(p, "src")
		fis, err := ioutil.ReadDir(src)
		if err == nil {
			for _, fi := range fis {
				if fi.IsDir() {
					anchors = append(anchors, fi.Name())
				}
			}
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(anchors))) // i.e. greedy; e.g. "foobar" before "foo"
}

func AbstractPath(p string) string {
	p = strings.TrimPrefix(p, "_")

	if ss := strings.SplitN(p, "src/", 2); len(ss) == 2 {
		return ss[1]
	}

	for _, a := range anchors {
		if ss := strings.SplitN(p, a+"/", 2); len(ss) == 2 {
			return filepath.Join(a, ss[1])
		}
	}

	return p
}
