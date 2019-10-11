## Golang logging library

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AndrewDonelson/golog) [![build](https://img.shields.io/travis/AndrewDonelson/golog.svg?style=flat)](https://travis-ci.org/AndrewDonelson/golog) [![Coverage Status](https://coveralls.io/repos/github/AndrewDonelson/golog/badge.svg?branch=develop)](https://coveralls.io/github/AndrewDonelson/golog?branch=develop)

Package logging implements a logging infrastructure for Go. Its output format
is customizable and supports different logging backends like syslog, file and
memory. Multiple backends can be utilized with different log levels per backend
and logger.

## Example

Let's have a look at an [example](examples/example.go) which demonstrates most
of the features found in this library.

[![Example Output](examples/example.png)](examples/example.go)

```go
package main

import (
	"os"
	"github.com/AndrewDonelson/golog"
)

var log = golog.MustGetLogger("example")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = golog.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// Password is just an example type implementing the Redactor interface. Any
// time this is logged, the Redacted() function will be called.
type Password string

func (p Password) Redacted() interface{} {
	return golog.Redact(string(p))
}

func main() {
	// For demo purposes, create two backend for os.Stderr.
	backend1 := golog.NewLogBackend(os.Stderr, "", 0)
	backend2 := golog.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := golog.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := golog.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(golog.ERROR, "")

	// Set the backends to be used.
	golog.SetBackend(backend1Leveled, backend2Formatter)

	log.Debugf("debug %s", Password("secret"))
	log.Success("success")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
}
```

## Installing

### Using *go get*

    $ go get github.com/AndrewDonelson/golog

After this command *go-logging* is ready to use. Its source will be in:

    $GOPATH/src/pkg/github.com/AndrewDonelson/golog

You can use `go get -u` to update the package.

## Documentation

For docs, see http://godoc.org/github.com/AndrewDonelson/golog or run:

    $ godoc github.com/AndrewDonelson/golog

## Additional resources

* [wslog](https://godoc.org/github.com/cryptix/exp/wslog) -- exposes log messages through a WebSocket.
