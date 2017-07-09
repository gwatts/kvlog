package kvlog_test

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gwatts/kvlog"
)

func ExampleWithConstantField() {
	f := kvlog.New(
		kvlog.WithConstantField("commit", "abcd1234"))

	result, _ := f.Format(&log.Entry{
		Time:  time.Date(2017, 1, 2, 12, 0, 0, 0, time.UTC),
		Level: log.InfoLevel,
		Data: log.Fields{
			"msg_count": 1,
		},
	})
	fmt.Println(string(result))

	// Output: 2017-01-02T12:00:00.000Z ll="info" commit="abcd1234" msg_count=1
}
