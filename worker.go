// Package golog Simple flexible go logging
// This file contains all code for the worker
package golog

import (
	"bytes"
	"io"
	"log"
)

// Worker class, Worker is a log object used to log messages and Color specifies
// if colored output is to be produced
type Worker struct {
	Minion      *log.Logger
	Environment Environment
	Color       ColorMode
	format      string
	timeFormat  string
	level       LogLevel
	function    string
}

// NewWorker Returns an instance of worker class, prefix is the string attached to every log,
// flag determine the log params, color parameters verifies whether we need colored outputs or not
func NewWorker(prefix string, flag int, color ColorMode, out io.Writer) *Worker {
	return &Worker{Minion: log.New(out, prefix, flag), Color: color, format: defFmt, timeFormat: defTimeFmt}
}

// SetFormat ...
func (w *Worker) SetFormat(format string) {
	w.format, w.timeFormat = parseFormat(format)
}

// SetLogLevel ...
func (w *Worker) SetLogLevel(level LogLevel) {
	w.level = level
}

// SetFunction sets the function name ofr the worker
func (w *Worker) SetFunction(name string) {
	w.function = name
}

// SetEnvironment is used to manually set the log environment to either development, testing or production
func (w *Worker) SetEnvironment(env Environment) {
	if w.Environment != EnvTesting {
		w.Environment = env
	}

	if env == EnvTesting {
		// set for testing
		w.level = InfoLevel
		w.format = defFmt
		return
	} else if env == EnvQuality {
		// set for qa
		w.level = InfoLevel
		w.format = defFmt
		return
	} else if env == EnvDevelopment {
		// set for developer
		w.level = DebugLevel
		w.format = defDevelopmentFmt
		return
	}

	// set for production
	w.level = ErrorLevel
	w.format = defProductionFmt
}

// SetOutput is used to manually set the output to send log data
func (w *Worker) SetOutput(out io.Writer) {
	w.Minion.SetOutput(out)
}

// Log Function of Worker class to log a string based on level
func (w *Worker) Log(level LogLevel, calldepth int, info *Info) error {

	info.Function = w.function

	if w.level < level {
		return nil
	}

	if w.Color == ClrAuto || w.Color == ClrEnabled {
		buf := &bytes.Buffer{}
		buf.Write([]byte(colors[level]))
		buf.Write([]byte(info.Output(w.format)))
		buf.Write([]byte("\033[0m"))
		return w.Minion.Output(calldepth+1, buf.String())
	}

	return w.Minion.Output(calldepth+1, info.Output(w.format))
}
