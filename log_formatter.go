// Copyright 2017 Gareth Watts
// Licensed under an MIT license
// See the LICENSE file for details

/*
Package kvlog implement a key=value log formatter for the logrus
logging package.

It can optionally include the calling function name and line number,
include constant values in every line and promote certain primary keys
to the beginning of each line.  All other keys are sorted into alphabetical
order for easy scanning, with the human-readnable description at the end
of each line.

eg.

  2017-07-09T17:00:05.460Z ll="info" srcfnc="(*MessageProcessor).handleDelivery" action="deliver_msg" status="ok" msg_count=1 _msg="delivered message ok"


This provides a format that's human-readable, yet automatically extracted by
tools such as Splunk.
*/
package kvlog

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	defaultStackDepth = 5
)

// Config represents a configuration function to be passed to New.
type Config func(kvf *Formatter)

// WithPrimaryFields specifies a number of field names that should always
// appear first in each log entry (if set).  The remaining fields will
// be sorted into alphabetical order.  This can be useful for fields which
// always appear that need to be seen quickly such as "status" or "action".
func WithPrimaryFields(field ...string) Config {
	return func(kvf *Formatter) {
		kvf.primaryFields = append([]string{}, field...)
	}
}

// WithConstantField specifies a field name and value that should be included
// in every log entry before any others (including primary fields).
func WithConstantField(key string, value interface{}) Config {
	return func(kvf *Formatter) {
		var buf bytes.Buffer
		kvf.emit(&buf, key, value, 0)
		kvf.constantFields = append(kvf.constantFields, buf.Bytes())
	}
}

// IncludeCaller causes the Formatter to include the calling function name
// in each log entry.
func IncludeCaller() Config {
	return func(kvf *Formatter) {
		kvf.includeCaller = true
	}
}

// Formatter emits plain text log lines with k="v" pairs.
type Formatter struct {
	primaryFields  []string
	constantFields [][]byte
	includeCaller  bool
	calcDepthOnce  sync.Once
	stackDepth     int
}

// New creates a new Formatter.
func New(cfgs ...Config) *Formatter {
	kvf := new(Formatter)
	for _, cfg := range cfgs {
		cfg(kvf)
	}
	return kvf
}

// Format a single log entry into a plain text log line.
func (cf *Formatter) Format(entry *log.Entry) ([]byte, error) {
	var buf bytes.Buffer

	cf.emitTimestamp(&buf, entry.Time)
	cf.emitLogLevel(&buf, entry.Level)
	if cf.includeCaller {
		cf.emitCaller(&buf)
	}

	for _, f := range cf.constantFields {
		buf.Write(f)
	}

	var skip map[string]struct{}
	if len(cf.primaryFields) > 0 {
		skip = make(map[string]struct{})
		for _, k := range cf.primaryFields {
			if v, ok := entry.Data[k]; ok {
				skip[k] = struct{}{}
				cf.emit(&buf, k, v, 0)
			}
		}
	}

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		if _, ok := skip[k]; ok {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cf.emit(&buf, k, entry.Data[k], 0)
	}

	if entry.Message != "" {
		cf.emit(&buf, "_msg", entry.Message, 0)
	}

	buf.Write([]byte("\n"))

	return buf.Bytes(), nil
}

func (cf *Formatter) emitTimestamp(b *bytes.Buffer, t time.Time) {
	buf := make([]byte, 0, 20)

	year, month, day := t.UTC().Date()
	hour, min, sec := t.UTC().Clock()
	ms := t.Nanosecond() / int(time.Millisecond)
	buf = itoa(buf, year, 4)
	buf = append(buf, '-')
	buf = itoa(buf, int(month), 2)
	buf = append(buf, '-')
	buf = itoa(buf, day, 2)
	buf = append(buf, 'T')
	buf = itoa(buf, hour, 2)
	buf = append(buf, ':')
	buf = itoa(buf, min, 2)
	buf = append(buf, ':')
	buf = itoa(buf, sec, 2)
	buf = append(buf, '.')
	buf = itoa(buf, ms, 3)
	buf = append(buf, 'Z')

	b.Write(buf)
}

func (cf *Formatter) emit(b *bytes.Buffer, k string, v interface{}, n int) {
	if v, ok := v.(Loggable); ok {
		kvs := v.LogValues()
		keys := make([]string, 0, len(kvs))
		for k := range kvs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, sk := range keys {
			cf.emit(b, k+sk, kvs[sk], n+1)
		}
		return
	}

	if n > -1 {
		b.Write([]byte{' '})
	}

	b.Write([]byte(k))
	b.Write([]byte{'='})

	switch data := v.(type) {
	case fmt.Stringer:
		fmt.Fprintf(b, "%+q", data)

	case string:
		fmt.Fprintf(b, "%+q", data)

	case *string:
		if data == nil {
			b.Write([]byte("<nil>"))
		} else {
			fmt.Fprintf(b, "%+q", *data)
		}

	case error:
		fmt.Fprintf(b, "%+q", data.Error())

	case []byte:
		fmt.Fprintf(b, "%+q", data)

	case Marshaler:
		b.Write([]byte(data.MarshalLogValue()))

	default:
		fmt.Fprintf(b, "%v", data)
	}
}

func (cf *Formatter) emitLogLevel(b *bytes.Buffer, level log.Level) {
	fmt.Fprintf(b, " ll=%q", level)
}

func (cf *Formatter) findCaller() (string, int) {
	callers := make([]uintptr, 10)
	runtime.Callers(3, callers) // set to 1 to skip Callers itself

	callingPackage := ""
	thispkg, _ := pkgnameForPC(callers[0])
	root := runtime.GOROOT()

	for _, pc := range callers {
		f := runtime.FuncForPC(pc)
		if f == nil {
			continue
		}
		pkg, funcname := pkgname(f.Name())
		fn, _ := f.FileLine(pc)

		switch {
		case pkg == thispkg:
		case callingPackage != "" && pkg == callingPackage:
		case strings.HasPrefix(fn, root): // stdlib
		case callingPackage == "":
			callingPackage = pkg
		default:
			_, line := f.FileLine(pc)
			return funcname, line
		}
	}
	return "", -1
}

func (cf *Formatter) emitCaller(b *bytes.Buffer) {
	name, line := cf.findCaller()
	if name == "" {
		b.Write([]byte(" srcfnc=\"unknown\""))
		return
	}

	fmt.Fprintf(b, " srcfnc=%q srcline=%d", name, line)
}

// Marshaler is the interface implemented by types that can marshal their own
// value into a log-friendly format.
//
// Values returned by types implementing this interface do not have newlines
// stripped, nor are their values quoted; they are included verbatim.
type Marshaler interface {
	MarshalLogValue() string
}

// RawLogString returns a LogValue so that newlines, etc
// are not stripped and the string is not quoted.
type RawLogString string

// MarshalLogValue implements the Marshaler interface.
func (s RawLogString) MarshalLogValue() string {
	return string(s)
}

var _ Marshaler = RawLogString("") // assert that RawLogString implements the Marshaler interface.

// Loggable is the interface implemented by types that contain multiple k=v
// values that need to be logged.
//
// Each key in the returned map will be prefixed with the value's key and then
// merged into the remaining keys for the log entry.  Each value will be
// formatted as any other value (each value may also implement the Marshaler
// interface, for example).
type Loggable interface {
	LogValues() map[string]interface{}
}

// lifted from log.go
// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf []byte, i int, wid int) []byte {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	buf = append(buf, b[bp:]...)
	return buf
}
