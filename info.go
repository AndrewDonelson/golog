// Package golog Simple flexible go logging
// this file contains all the code for Info
package golog

import (
	"fmt"
	"strings"
)

// Info class, Contains all the info on what has to logged, time is the current time, Module is the specific module
// For which we are logging, level is the state, importance and type of message logged,
// Message contains the string to be logged, format is the format of string to be passed to sprintf
type Info struct {
	ID       uint64
	Time     string
	Module   string
	Function string
	Level    LogLevel
	Line     int
	Filename string
	Message  string
	//format   string
}

// Output Returns a proper string to be outputted for a particular info
func (r *Info) Output(format string) string {
	msg := fmt.Sprintf(format,
		r.ID,               // %[1] // %{id}
		r.Time,             // %[2] // %{time[:fmt]}
		r.Module,           // %[3] // %{module}
		r.Function,         // %[4] // %{function}
		r.Filename,         // %[5] // %{filename}
		r.Line,             // %[6] // %{line}
		r.logLevelString(), // %[7] // %{level}
		r.Message,          // %[8] // %{message}
	)
	// Ignore printf errors if len(args) > len(verbs)
	if i := strings.LastIndex(msg, "%!(EXTRA"); i != -1 {
		return msg[:i]
	}
	return msg
}

// logLevelString Returns the loglevel as string
func (r *Info) logLevelString() string {
	logLevels := [...]string{
		"CRITICAL",
		"ERROR",
		"SUCCESS",
		"WARNING",
		"NOTICE",
		"INFO",
		"DEBUG",
	}
	return logLevels[r.Level-1]
}
