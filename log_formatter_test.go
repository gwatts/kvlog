// Copyright 2017 Gareth Watts
// Licensed under an MIT license
// See the LICENSE file for details

package kvlog_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/gwatts/kvlog"
)

var testTime = parseTime("2017-02-13T12:13:45Z")

func parseTime(tm string) time.Time {
	t, err := time.Parse(time.RFC3339, "2017-02-13T12:13:45Z")
	if err != nil {
		panic(fmt.Sprintf("time parse failed: %v", err))
	}
	return t
}

var fmtTests = []struct {
	name     string
	order    []string
	entry    *log.Entry
	expected string
}{
	{"simple", nil, &log.Entry{
		Time:  testTime,
		Level: log.InfoLevel,
		Data: log.Fields{
			"field1":      "str with spaces",
			"field2":      123,
			"field-three": errors.New("test error"),
		},
	}, `2017-02-13T12:13:45.000Z ll="info" field-three="test error" field1="str with spaces" field2=123`},
	{"ordered", []string{"rand1", "another", "rand2", "rand3"}, &log.Entry{
		Time:  testTime,
		Level: log.InfoLevel,
		Data: log.Fields{
			"field1":  "str with spaces",
			"field2":  123,
			"another": "another-field",
			"rand1":   "foobar",
		},
	}, `2017-02-13T12:13:45.000Z ll="info" rand1="foobar" another="another-field" field1="str with spaces" field2=123`},
	{"with-message", nil, &log.Entry{
		Time:    testTime,
		Level:   log.InfoLevel,
		Message: "test message",
		Data: log.Fields{
			"field1": "value1",
		},
	}, `2017-02-13T12:13:45.000Z ll="info" field1="value1" _msg="test message"`,
	},
}

func TestFormatter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	for _, test := range fmtTests {
		cf := New(
			WithPrimaryFields(test.order...))
		result, err := cf.Format(test.entry)
		require.Nil(err, test.name+" should not error")
		assert.Equal(test.expected, strings.TrimSpace(string(result)), test.name+" should match")
	}
}

func TestConstantField(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cf := New(
		WithConstantField("field1", "value1"),
		WithConstantField("field2", 123),
	)
	result, err := cf.Format(&log.Entry{
		Time:  testTime,
		Level: log.InfoLevel,
		Data: log.Fields{
			"varfield1": "vf1",
		},
	})
	require.Nil(err, "Should not error")
	expected := `2017-02-13T12:13:45.000Z ll="info" field1="value1" field2=123 varfield1="vf1"`
	assert.Equal(expected, strings.TrimSpace(string(result)))
}

func TestLogEmitter(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	cf := New(IncludeCaller())
	logger := &log.Logger{
		Out:       &buf,
		Formatter: cf,
		Level:     log.DebugLevel,
	}
	logger.Info(log.WithFields(log.Fields{
		"field1": "value1",
	}))
	result := strings.TrimSpace(buf.String())
	assert.Contains(result, "TestLogEmitter")
}

func BenchmarkEmitter(b *testing.B) {
	var buf bytes.Buffer
	cf := New(IncludeCaller())
	logger := &log.Logger{
		Out:       &buf,
		Formatter: cf,
		Level:     log.DebugLevel,
	}
	fields := log.Fields{"f1": "b"}

	for i := 0; i < b.N; i++ {
		logger.Info(fields)
	}
}

func BenchmarkNoEmitter(b *testing.B) {
	var buf bytes.Buffer
	cf := New()
	logger := &log.Logger{
		Out:       &buf,
		Formatter: cf,
		Level:     log.DebugLevel,
	}
	fields := log.Fields{"f1": "b"}

	for i := 0; i < b.N; i++ {
		logger.Info(fields)
	}
}
