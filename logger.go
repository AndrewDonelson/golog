// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package golog implements a golog infrastructure for Go. It supports
// different golog backends like syslog, file and memory. Multiple backends
// can be utilized with different log levels per backend and logger.
package golog

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Redactor is an interface for types that may contain sensitive information
// (like passwords), which shouldn't be printed to the log. The idea was found
// in relog as part of the vitness project.
type Redactor interface {
	Redacted() interface{}
}

// Redact returns a string of * having the same length as s.
func Redact(s string) string {
	return strings.Repeat("*", len(s))
}

var (
	// Sequence number is incremented and utilized for all log records created.
	sequenceNo uint64

	// timeNow is a customizable for testing purposes.
	timeNow = time.Now
)

// Record represents a log record and contains the timestamp when the record
// was created, an increasing id, filename and line and finally the actual
// formatted log line.
type Record struct {
	ID     uint64
	Time   time.Time
	Module string
	Level  Level
	Args   []interface{}

	// message is kept as a pointer to have shallow copies update this once
	// needed.
	message    *string
	fmt        *string
	formatter  Formatter
	formatted  string
	prefix     string
	dumpPrefix string
	isDump     bool
}

// Formatted returns the formatted log record string.
func (r *Record) Formatted(calldepth int) string {
	if r.formatted == "" {
		var buf bytes.Buffer
		r.formatter.Format(calldepth+1, r, &buf)
		r.formatted = buf.String()
		if r.isDump {
			r.formatted = r.dumpPrefix + r.formatted
		}
	}
	return r.formatted
}

// Message returns the log record message.
func (r *Record) Message() string {
	if r.message == nil {
		// Redact the arguments that implements the Redactor interface
		for i, arg := range r.Args {
			if redactor, ok := arg.(Redactor); ok == true {
				r.Args[i] = redactor.Redacted()
			}
		}
		var buf bytes.Buffer
		if r.fmt != nil {
			fmt.Fprintf(&buf, *r.fmt, r.Args...)
		} else {
			// use Fprintln to make sure we always get space between arguments
			fmt.Fprintln(&buf, r.Args...)
			buf.Truncate(buf.Len() - 1) // strip newline
		}
		msg := buf.String()
		r.message = &msg
	}
	return *r.message
}

// A ring buffer is used in the every logger instance to store several last records
func newRingBuffer(capacity int) *Ring {
	r := &Ring{}
	r.SetCapacity(capacity)
	return r
}

// Logger is the actual logger which creates log records based on the functions
// called and passes them to the underlying golog backend.
type Logger struct {
	Module      string
	backend     Backend
	haveBackend bool

	// ExtraCallDepth can be used to add additional call depth when getting the
	// calling function. This is normally used when wrapping a logger.
	ExtraCalldepth int
	Context        string

	//for dumping
	enableDumping bool
	triggerLevel  Level
	records       *Ring
	capacity      int
	mutex         *sync.Mutex
	formatter     Formatter
	level         Level
	ctxSep        string
	dumpPrefix    string
}

// SetFormatter ...
func (l *Logger) SetFormatter(fmt Formatter) {
	l.formatter = fmt
}

// SetDumpBehavior ...
func (l *Logger) SetDumpBehavior(enable bool, lvl Level, dumpPrefix string) {
	l.enableDumping = enable
	l.triggerLevel = lvl
	l.dumpPrefix = dumpPrefix
}

// WithContext builds new logger with new ring buffer inside but with inherited context
func (l *Logger) WithContext(context string) *Logger {
	newLogger := *l
	newLogger.mutex = &sync.Mutex{}
	newLogger.records = newRingBuffer(l.capacity)
	newLogger.Context = l.Context + context + l.ctxSep
	return &newLogger
}

// WithLevel ...
func (l *Logger) WithLevel(level Level) *Logger {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// SetBackend overrides any previously defined backend for this logger.
func (l *Logger) SetBackend(backend Backend) {
	l.backend = backend
	l.haveBackend = true
}

// SetLevel ...
func (l *Logger) SetLevel(lvl Level) {
	l.level = lvl
}

// TODO call NewLogger and remove MustGetLogger?

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string, capacity int) (*Logger, error) {
	return &Logger{
		Module:  module,
		records: newRingBuffer(capacity),
		mutex:   &sync.Mutex{},
	}, nil
}

// MustGetLogger is like GetLogger but panics if the logger can't be created.
// It simplifies safe initialization of a global logger for eg. a package.
func MustGetLogger(module, contextSeparator string, capacity int) *Logger {
	logger, err := GetLogger(module, capacity)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
	logger.ctxSep = contextSeparator
	return logger
}

// Reset restores the internal state of the golog library.
func Reset() {
	// TODO make a global Init() method to be less magic? or make it such that
	// if there's no backends at all configured, we could use some tricks to
	// automatically setup backends based if we have a TTY or not.
	sequenceNo = 0
	timeNow = time.Now
}

// IsEnabledFor returns true if the logger is enabled for the given level.
func (l *Logger) IsEnabledFor(level Level) bool {
	return level <= l.level
}

func (l *Logger) checkAndDumpRecords(level Level) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if level <= l.triggerLevel {
		for {
			val := l.records.Dequeue()
			if val == nil {
				break
			}
			if l.haveBackend {
				l.backend.LogStr(level, 3+l.ExtraCalldepth, fmt.Sprintf("%v", val))
				// if msg, ok := Str; !ok {
				// 	panic(fmt.Sprintf("Severe implementation error: cannot cast %+v to string", val))
				// } else {
				// 	l.backend.LogStr(level, 3+l.ExtraCalldepth, msg)
				// }
			}
		}
	}
}

func (l *Logger) log(lvl Level, format *string, args ...interface{}) {

	// Create the golog record and pass it in to the backend
	record := &Record{
		ID:         atomic.AddUint64(&sequenceNo, 1),
		Time:       timeNow(),
		Module:     l.Module,
		Level:      lvl,
		fmt:        format,
		Args:       args,
		prefix:     l.Context,
		formatter:  l.formatter,
		dumpPrefix: l.dumpPrefix,
	}

	if l.enableDumping && lvl > l.triggerLevel && !l.IsEnabledFor(lvl) {
		l.mutex.Lock()
		str := l.dumpPrefix + record.Formatted(2+l.ExtraCalldepth)
		l.records.Enqueue(str)
		l.mutex.Unlock()
	}

	if !l.IsEnabledFor(lvl) {
		return
	}

	// TODO use channels to fan out the records to all backends?
	// TODO in case of errors, do something (tricky)

	// calldepth=2 brings the stack up to the caller of the level
	// methods, Info(), Fatal(), etc.
	// ExtraCallDepth allows this to be extended further up the stack in case we
	// are wrapping these methods, eg. to expose them package level
	if l.haveBackend {
		l.backend.Log(lvl, 2+l.ExtraCalldepth, record)
		l.checkAndDumpRecords(lvl)
		return
	}
}

// Printf ...
func (l *Logger) Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	if l.haveBackend {
		l.backend.LogStr(DEBUG, 1+l.ExtraCalldepth, str)
		return
	}
}

// Print ...
func (l *Logger) Print(args ...interface{}) {
	str := fmt.Sprint(args...)
	if l.haveBackend {
		l.backend.LogStr(DEBUG, 1+l.ExtraCalldepth, str)
		return
	}
}

// Fatal is equivalent to l.Critical(fmt.Sprint()) followed by a call to os.Exit(1).
func (l *Logger) Fatal(args ...interface{}) {
	l.log(CRITICAL, nil, args...)
	os.Exit(1)
}

// Fatalf is equivalent to l.Critical followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(CRITICAL, &format, args...)
	os.Exit(1)
}

// Panic is equivalent to l.Critical(fmt.Sprint()) followed by a call to panic().
func (l *Logger) Panic(args ...interface{}) {
	l.log(CRITICAL, nil, args...)
	panic(fmt.Sprint(args...))
}

// Panicf is equivalent to l.Critical followed by a call to panic().
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log(CRITICAL, &format, args...)
	panic(fmt.Sprintf(format, args...))
}

// Critical logs a message using CRITICAL as log level.
func (l *Logger) Critical(args ...interface{}) {
	l.log(CRITICAL, nil, args...)
}

// Criticalf logs a message using CRITICAL as log level.
func (l *Logger) Criticalf(format string, args ...interface{}) {
	l.log(CRITICAL, &format, args...)
}

// Error logs a message using ERROR as log level.
func (l *Logger) Error(args ...interface{}) {
	l.log(ERROR, nil, args...)
}

// Errorf logs a message using ERROR as log level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, &format, args...)
}

// Warning logs a message using WARNING as log level.
func (l *Logger) Warning(args ...interface{}) {
	l.log(WARNING, nil, args...)
}

// Warningf logs a message using WARNING as log level.
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.log(WARNING, &format, args...)
}

// Notice logs a message using NOTICE as log level.
func (l *Logger) Notice(args ...interface{}) {
	l.log(NOTICE, nil, args...)
}

// Noticef logs a message using NOTICE as log level.
func (l *Logger) Noticef(format string, args ...interface{}) {
	l.log(NOTICE, &format, args...)
}

// Info logs a message using INFO as log level.
func (l *Logger) Info(args ...interface{}) {
	l.log(INFO, nil, args...)
}

// Infof logs a message using INFO as log level.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, &format, args...)
}

// Debug logs a message using DEBUG as log level.
func (l *Logger) Debug(args ...interface{}) {
	l.log(DEBUG, nil, args...)
}

// Debugf logs a message using DEBUG as log level.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, &format, args...)
}

// Success logs a message using SUCCESS as log level.
func (l *Logger) Success(args ...interface{}) {
	l.log(SUCCESS, nil, args...)
}

// Successf logs a message using SUCCESS as log level.
func (l *Logger) Successf(format string, args ...interface{}) {
	l.log(SUCCESS, &format, args...)
}

func init() {
	Reset()
}
