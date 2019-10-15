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

	//log, err := New("test", 0, &buf)
	log, err := NewLogger(nil)
	if err != nil {
		t.Error(err) // Check for error
		return
	}
	log.Options.Out = &buf

	// Get current function name
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	//fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)

	_, filename, line, _ := runtime.Caller(1)
	filename = path.Base(filename)
	info := &Info{
		ID:       atomic.AddUint64(&logNo, 1),
		Time:     time.Now().Format(log.worker.timeFormat),
		Module:   log.Module,
		Function: frame.Function,
		Level:    WarningLevel,
		Message:  "Hello World!",
		Filename: filename,
		Line:     line,
		//format:   formatString,
	}
	err = log.worker.Log(WarningLevel, 2, info)
	if err != nil {
		t.Error(err)
		return
	}

	want := fmt.Sprintf("[33mtest %s WARNING ▶ testing.tRunner ▶ Hello World![0m\n", time.Now().Format("2006-01-02 15:04:05"))
	have := buf.String()
	if have != want {
		t.Errorf("\nWant: %sHave: %s", want, have)
	}
	buf.Reset()
}
