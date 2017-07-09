package kvlog_test

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gwatts/kvlog"
)

type AgeRange struct {
	Youngest int
	Oldest   int
}

func (ar AgeRange) MarshalLogValue() string {
	return fmt.Sprintf(`"%d-%d"`, ar.Youngest, ar.Oldest)
}

func ExampleMarshaler() {
	ar := AgeRange{
		Youngest: 18,
		Oldest:   93,
	}

	f := kvlog.New()

	result, _ := f.Format(&log.Entry{
		Time:  time.Date(2017, 1, 2, 12, 0, 0, 0, time.UTC),
		Level: log.InfoLevel,
		Data: log.Fields{
			"age_range": ar,
		},
	})
	fmt.Println(string(result))

	// Output: 2017-01-02T12:00:00.000Z ll="info" age_range="18-93"
}
