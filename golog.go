// Package golog Simple flexible go logging
// This file contains all the code for the main logger
package golog

// Import packages
import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	// Default format of log message
	// %[1] // %{id}
	// %[2] // %{time[:fmt]}
	// %[3] // %{module}
	// %[4] // %{function}
	// %[5] // %{filename}
	// %[6] // %{line}
	// %[7] // %{level}
	// %[8] // %{message}
	defProductionFmt  = "%.16[3]s %.19[2]s %.3[7]s ▶ %[8]s"
	defDevelopmentFmt = "%.16[3]s %.19[2]s %.8[7]s ▶ %[4]s ▶ %[8]s"

	// Error, Fatal, Critical Format
	//defErrorFmt = "%.16[3]s %.19[2]s %.8[7]s ▶ %[8]s\n▶ %[5]s:%[6]d-%[4]s"
)

var (
	// Map for the various codes of colors
	colors map[LogLevel]string

	// Map from format's placeholders to printf verbs
	phfs map[string]string

	// Contains color strings for stdout
	logNo uint64

	defFmt = "#%[1]d %.19[2]s %[5]s:%[6]d ▶ %.3[7]s %[8]s"

	// Default format of time
	defTimeFmt = "2006-01-02 15:04:05"
)

// LogLevel type
type LogLevel int

// Color numbers for stdout
const (
	Black = (iota + 30)
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Log Level
const (
	CriticalLevel LogLevel = iota + 1 // Magneta 	35
	ErrorLevel                        // Red 		31
	SuccessLevel                      // Green 		32
	WarningLevel                      // Yellow 	33
	NoticeLevel                       // Cyan 		36
	InfoLevel                         // White 		37
	DebugLevel                        // Blue 		34
)

// Logger class that is an interface to user to log messages, Module is the module for which we are testing
// worker is variable of Worker class that is used in bottom layers to log the message
type Logger struct {
	Options Options
	Module  string
	worker  *Worker
}

// init is called by NewLogger to detect running conditions and set all defaults
func (l *Logger) init() {
	if flag.Lookup("test.v") != nil {
		println("Environment: Testing")
		l.SetEnvironment(-1)
	} else {
		be := os.Getenv("BUILD_ENV")
		if be == "dev" {
			println("Environment: Development")
			l.SetEnvironment(2)
		} else if be == "qa" {
			println("Environment: Quality Assurance")
			l.SetEnvironment(1)
		} else {
			println("Environment: Production")
			l.SetEnvironment(0)
		}
	}

	initColors()
	initFormatPlaceholders()
}

// NewLogger creates and returns new logger for the given model & environment
// module is the specific module for which we are logging
// environment overrides detected environment (if -1)
// color defines whether the output is to be colored or not, out is instance of type io.Writer defaults
// to os.Stderr
func NewLogger(opts *Options) (*Logger, error) {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	if opts.Out == nil {
		opts.Out = os.Stderr
	}

	if len(opts.Module) <= 3 {
		opts.Module = "unknown"
	}

	newWorker := NewWorker("", 0, opts.UseColor, opts.Out)
	l := &Logger{Module: opts.Module, worker: newWorker}
	l.init()
	l.Options = *opts
	newWorker.SetEnvironment(opts.Environment)
	return l, nil

}

// SetDefaultFormat ...
func SetDefaultFormat(format string) {
	defFmt, defTimeFmt = parseFormat(format)
}

// SetFormat ...
func (l *Logger) SetFormat(format string) {
	l.worker.SetFormat(format)
}

// SetLogLevel ...
func (l *Logger) SetLogLevel(level LogLevel) {
	l.worker.SetLogLevel(level)
}

// SetFunction sets the function name of the logger
func (l *Logger) SetFunction(name string) {
	l.worker.SetFunction(name)
}

// SetEnvironment is used to manually set the log environment to either development, testing or production
func (l *Logger) SetEnvironment(env int) {
	l.Options.Environment = env
	l.worker.SetEnvironment(env)
}

// SetOutput is used to manually set the output to send log data
func (l *Logger) SetOutput(out io.Writer) {
	l.Options.Out = out
	l.worker.SetOutput(out)
}

// Log The log command is the function available to user to log message,
// lvl specifies the degree of the message the user wants to log, message
// is the info user wants to log
func (l *Logger) Log(lvl LogLevel, message string) {
	l.logInternal(lvl, message, 2)
}

// logInternal ...
func (l *Logger) logInternal(lvl LogLevel, message string, pos int) {
	_, filename, line, _ := runtime.Caller(pos)
	filename = path.Base(filename)
	info := &Info{
		ID:       atomic.AddUint64(&logNo, 1),
		Time:     time.Now().Format(l.worker.timeFormat),
		Module:   l.Module,
		Level:    lvl,
		Message:  message,
		Filename: filename,
		Line:     line,
		//format:   formatString,
	}
	err := l.worker.Log(lvl, 2, info)
	if err != nil {
		panic(err)
	}
}

// Fatal is just like func l.Critical logger except that it is followed by exit to program
func (l *Logger) Fatal(message string) {
	l.logInternal(CriticalLevel, message, 2)
	os.Exit(1)
}

// Fatalf is just like func l.CriticalF logger except that it is followed by exit to program
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.logInternal(CriticalLevel, fmt.Sprintf(format, a...), 2)
	os.Exit(1)
}

// Panic is just like func l.Critical except that it is followed by a call to panic
func (l *Logger) Panic(message string) {
	l.logInternal(CriticalLevel, message, 2)
	panic(message)
}

// Panicf is just like func l.CriticalF except that it is followed by a call to panic
func (l *Logger) Panicf(format string, a ...interface{}) {
	l.logInternal(CriticalLevel, fmt.Sprintf(format, a...), 2)
	panic(fmt.Sprintf(format, a...))
}

// Critical logs a message at a Critical Level
func (l *Logger) Critical(message string) {
	l.logInternal(CriticalLevel, message, 2)
}

// Criticalf logs a message at Critical level using the same syntax and options as fmt.Printf
func (l *Logger) Criticalf(format string, a ...interface{}) {
	l.logInternal(CriticalLevel, fmt.Sprintf(format, a...), 2)
}

// Error logs a message at Error level
func (l *Logger) Error(message string) {
	l.logInternal(ErrorLevel, message, 2)
}

// Errorf logs a message at Error level using the same syntax and options as fmt.Printf
func (l *Logger) Errorf(format string, a ...interface{}) {
	l.logInternal(ErrorLevel, fmt.Sprintf(format, a...), 2)
}

// Success logs a message at Success level
func (l *Logger) Success(message string) {
	l.logInternal(SuccessLevel, message, 2)
}

// Successf logs a message at Success level using the same syntax and options as fmt.Printf
func (l *Logger) Successf(format string, a ...interface{}) {
	l.logInternal(SuccessLevel, fmt.Sprintf(format, a...), 2)
}

// Warning logs a message at Warning level
func (l *Logger) Warning(message string) {
	l.logInternal(WarningLevel, message, 2)
}

// Warningf logs a message at Warning level using the same syntax and options as fmt.Printf
func (l *Logger) Warningf(format string, a ...interface{}) {
	l.logInternal(WarningLevel, fmt.Sprintf(format, a...), 2)
}

// Notice logs a message at Notice level
func (l *Logger) Notice(message string) {
	l.logInternal(NoticeLevel, message, 2)
}

// Noticef logs a message at Notice level using the same syntax and options as fmt.Printf
func (l *Logger) Noticef(format string, a ...interface{}) {
	l.logInternal(NoticeLevel, fmt.Sprintf(format, a...), 2)
}

// Info logs a message at Info level
func (l *Logger) Info(message string) {
	l.logInternal(InfoLevel, message, 2)
}

// Infof logs a message at Info level using the same syntax and options as fmt.Printf
func (l *Logger) Infof(format string, a ...interface{}) {
	l.logInternal(InfoLevel, fmt.Sprintf(format, a...), 2)
}

// Debug logs a message at Debug level
func (l *Logger) Debug(message string) {
	l.logInternal(DebugLevel, message, 2)
}

// Debugf logs a message at Debug level using the same syntax and options as fmt.Printf
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.logInternal(DebugLevel, fmt.Sprintf(format, a...), 2)
}

// StackAsError Prints this goroutine's execution stack as an error with an optional message at the begining
func (l *Logger) StackAsError(message string) {
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	l.logInternal(ErrorLevel, message+Stack(), 2)
}

// StackAsCritical Prints this goroutine's execution stack as critical with an optional message at the begining
func (l *Logger) StackAsCritical(message string) {
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	l.logInternal(CriticalLevel, message+Stack(), 2)
}

// Stack Returns a string with the execution stack for this goroutine
func Stack() string {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	return string(buf)
}
