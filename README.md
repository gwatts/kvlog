# Key=Value Log Formatter for Logrus

[![GoDoc](https://godoc.org/github.com/gwatts/kvlog?status.svg)](https://godoc.org/github.com/gwatts/kvlog)
[![Build Status](https://travis-ci.org/gwatts/kvlog.svg?branch=master)](https://travis-ci.org/gwatts/kvlog)

This package provides a text formatting type for Logrus.  It is similar to the
text formatter that comes with logrus, but is focused on compatibility with
logging systems such as Splunk.

It provides a number of features

* All fields logged as key=value format (which Splunk automatically extracts).
* Fields are sorted into order.
* Important/primary fields can be pinned to the start of each log entry
so they're easy to spot.
* Constant fields can be defined within the formatter.  For example, a build
commit hash can be included in every log entry automatically.
* All string types are wrapped in quotes automatically.
* Types can define their own marshaler for custom behaviour
* Compound types can return multiple key/value pairs by implementing a
Loggable interface.
* The calling function can optionally be included in every log entry.


Example usage:

```golang
log.SetFormatter(
		kvlog.New(
				// include the calling function name
				kvlog.IncludeCaller(), 

				// ensure action and status always appear first
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

// Output:
//  2017-01-02T12:00:00.000Z ll="info" srcfnc="Example" srcline=29 action="user_login" status="ok" active_sessions=4 email="joe@example.com" username="joe_user" _msg="User logged in"
```
