package kvlog_test

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gwatts/kvlog"
)

type Timing struct {
	Min    int
	Max    int
	Median int
}

func (t Timing) LogValues() map[string]interface{} {
	return map[string]interface{}{
		".min_ms":    t.Min,
		".max_ms":    t.Max,
		".median_ms": t.Median,
	}
}

func ExampleLoggable() {
	t := Timing{
		Min:    5,
		Max:    93,
		Median: 30,
	}

	f := kvlog.New()

	result, _ := f.Format(&log.Entry{
		Time:  time.Date(2017, 1, 2, 12, 0, 0, 0, time.UTC),
		Level: log.InfoLevel,
		Data: log.Fields{
			"exec_times": t,
		},
	})
	fmt.Println(string(result))

	// Output: 2017-01-02T12:00:00.000Z ll="info" exec_times.max_ms=93 exec_times.median_ms=30 exec_times.min_ms=5
}
