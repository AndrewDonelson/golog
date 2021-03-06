package golog

import (
	"bytes"
	"fmt"
	"path"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewInfo(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(nil)
	log.SetOutput(&buf)
	log.SetLogLevel(DebugLevel)

	// Get current function name
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	_, filename, line, _ := runtime.Caller(1)
	filename = path.Base(filename)
	info := &Info{
		ID:       atomic.AddUint64(&logNo, 1),
		Time:     time.Now().Format(log.worker.timeFormat),
		Module:   log.Options.Module,
		Function: frame.Function,
		Level:    InfoLevel,
		Message:  "Hello World!",
		Filename: filename,
		Line:     line,
	}
	log.worker.Log(ErrorLevel, 2, info)

	// "[35munknown 2019-10-15 19:20:51 INF - Hello World![0m"
	//want := fmt.Sprintf("[31m[unknown] #80 %s INFO Hello World![0m\n", time.Now().Format("2006-01-02 15:04:05"))
	want := fmt.Sprintf("[31m[000080] [unknown] INFO %s testing.go#909-testing.tRunner : Hello World![0m\n", time.Now().Format("2006-01-02 15:04:05"))
	have := buf.String()
	if have != want {
		t.Errorf("\nWant: %sHave: %s", want, have)
	}
	buf.Reset()
}
