package kvlog_test

import (
	"bytes"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/gwatts/kvlog"
)

func Example() {
	var buf bytes.Buffer

	log.SetOutput(&buf)
	log.SetFormatter(
		kvlog.New(
			kvlog.IncludeCaller(),
			kvlog.WithPrimaryFields("action", "status")))

	log.WithFields(log.Fields{
		"action":          "user_login",
		"status":          "ok",
		"username":        "joe_user",
		"email":           "joe@example.com",
		"active_sessions": 4,
	}).Info("User logged in")

	// replace the timestamp so the output is consistent
	output := "2017-01-02T12:00:00.000Z " + buf.String()[25:]
	fmt.Println(output)

	// Output: 2017-01-02T12:00:00.000Z ll="info" srcfnc="Example" srcline=29 action="user_login" status="ok" active_sessions=4 email="joe@example.com" username="joe_user" _msg="User logged in"
}
