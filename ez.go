// Copyright 2014 Alvaro J. Genial. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ez provides an easy, powerful way to define tests & benchmarks that are compatible with package `testing`.
package ez

var PathStyle = Absolute

type Style int

const (
	Absolute Style = iota
	Abstract
	Relative
	Truncate
)

func (s Style) String() string {
	switch s {
	case Absolute:
		return "Absolute"
	case Abstract:
		return "Abstract"
	case Relative:
		return "Relative"
	case Truncate:
		return "Truncate"
	default:
		return ""
	}
}
