// Copyright 2017 Gareth Watts
// Licensed under an MIT license
// See the LICENSE file for details

package kvlog

import (
	"runtime"
	"strings"
)

// attempt to extract the package name from a fully qualified function name
// won't work correctly if the package itself has a period in the name :-(
func pkgname(name string) (string, string) {
	fullName := name
	firstParen := strings.IndexByte(name, '(')
	if firstParen > 0 {
		name = name[:firstParen]
	}
	var termDot int
	lastSlash := strings.LastIndexByte(name, '/')
	if lastSlash > 0 {
		pos := strings.IndexByte(name[lastSlash:], '.')
		if pos == -1 {
			return "", ""
		}
		termDot = pos + lastSlash
	} else {
		termDot = strings.IndexByte(name, '.')
	}

	return name[:termDot], fullName[termDot+1:]
}

func pkgnameForPC(pc uintptr) (string, string) {
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "", ""
	}
	return pkgname(f.Name())
}
