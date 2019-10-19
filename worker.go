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
	environment Environment
	color       ColorMode
	format      string
	timeFormat  string
	level       LogLevel
	function    string
}

// NewWorker Returns an instance of worker class, prefix is the string attached to every log,
// flag determine the log params, color parameters verifies whether we need colored outputs or not
func NewWorker(prefix string, flag int, color ColorMode, out io.Writer) *Worker {
	return &Worker{Minion: log.New(out, prefix, flag), color: color, format: defFmt, timeFormat: defTimeFmt}
}

// UseJSONForProduction forces using JSON instead of log for production
func (w *Worker) UseJSONForProduction() {
	if w.environment == EnvProduction || w.environment == EnvTesting {
		w.format = FmtProductionJSON
	}
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

// GetEnvironment returns the currently set environment for the worker
func (w *Worker) GetEnvironment() Environment {
	return w.environment
}

// SetEnvironment is used to manually set the log environment to either development, testing or production
func (w *Worker) SetEnvironment(env Environment) {
	if w.environment != EnvTesting {
		w.environment = env
	}

	if env == EnvTesting {
		// set for testing
		w.level = InfoLevel
		w.format = defFmt
		w.color = ClrAuto
		return
	} else if env == EnvQuality {
		// set for qa
		w.level = InfoLevel
		w.format = defFmt
		w.color = ClrAuto
		return
	} else if env == EnvDevelopment {
		// set for developer
		w.level = DebugLevel
		w.format = FmtDevelopmentLog
		w.color = ClrAuto
		return
	}

	// set for production
	w.level = SuccessLevel
	w.format = FmtProductionLog
	w.color = ClrDisabled
}

// SetOutput is used to manually set the output to send log data
func (w *Worker) SetOutput(out io.Writer) {
	w.Minion.SetOutput(out)
}

// Log Function of Worker class to log a string based on level
func (w *Worker) Log(level LogLevel, calldepth int, info *Info) error {
	info.Function = w.function

	// Support RawLevel on any environment
	clr := w.color
	if level != RawLevel {
		if w.level < level {
			return nil
		}
	} else {
		clr = ClrDisabled
	}

	// Color for supported Levels
	if clr == ClrAuto || clr == ClrEnabled {
		buf := &bytes.Buffer{}
		buf.Write([]byte(colors[level]))
		buf.Write([]byte(info.Output(w.format)))
		buf.Write([]byte("\033[0m"))
		return w.Minion.Output(calldepth+1, buf.String())
	}

	// Regular no color output
	return w.Minion.Output(calldepth+1, info.Output(w.format))
}
