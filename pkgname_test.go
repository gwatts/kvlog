// Copyright 2017 Gareth Watts
// Licensed under an MIT license
// See the LICENSE file for details

package kvlog

import "testing"

var pkgNameTests = []struct {
	pkgname string
	fname   string
}{
	{"github.com/gwatts/kvlog", "github.com/gwatts/kvlog.(*KVFormatter).calcStackDepth"},
	{"github.com/gwatts/kvlog", "github.com/gwatts/kvlog.(*KVFormatter).(github.com/gwatts/kvlog.calcStackDepth)-fm"},
	{"sync", "sync.(*Once).Do"},
	{"github.com/Sirupsen/logrus", "github.com/Sirupsen/logrus.Entry.log"},
	{"testing", "testing.tRunner"},
	{"foo/bar", "foo/bar.Exported"},
}

func TestPkgName(t *testing.T) {
	for _, test := range pkgNameTests {
		result, _ := pkgname(test.fname)
		if result != test.pkgname {
			t.Errorf("input=%q expected=%q actual=%q", test.fname, test.pkgname, result)
		}
	}
}
