// Package golog Simple flexible go logging
package golog

// Import packages
import (
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
	Module string
	worker *Worker
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
	l.worker.level = level
}

// SetFunction sets the function name of the logger
func (l *Logger) SetFunction(name string) {
	l.worker.function = name
}

// NewLogger creates a new logger for the given model & environment
func NewLogger(module string, environment int) (*Logger, error) {
	var (
		color  int       = 1
		out    io.Writer = os.Stderr
		level  LogLevel  = ErrorLevel
		format string    = defProductionFmt
	)

	if len(module) <= 3 {
		panic("You must provide a name for the module (app, rpc, etc)")
	}

	if environment == 1 {
		// set for test (qa)
		level = InfoLevel
		format = defFmt
	} else if environment == 2 {
		// set for developer
		level = DebugLevel
		format = defDevelopmentFmt
	}

	newWorker := NewWorker("", 0, color, out)
	newWorker.SetLogLevel(level)
	newWorker.SetFormat(format)
	return &Logger{Module: module, worker: newWorker}, nil
}

// New Returns a new instance of logger class, module is the specific module for which we are logging
// , color defines whether the output is to be colored or not, out is instance of type io.Writer defaults
// to os.Stderr
func New(args ...interface{}) (*Logger, error) {
	//initColors()

	var (
		module string    = "DEFAULT"
		color  int       = 1
		out    io.Writer = os.Stderr
		level  LogLevel  = InfoLevel
	)

	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			module = t
		case int:
			color = t
		case io.Writer:
			out = t
		case LogLevel:
			level = t
		default:
			panic("logger: Unknown argument")
		}
	}
	newWorker := NewWorker("", 0, color, out)
	newWorker.SetLogLevel(level)
	return &Logger{Module: module, worker: newWorker}, nil
}

// Log The log command is the function available to user to log message,
// lvl specifies the degree of the message the user wants to log, message
// is the info user wants to log
func (l *Logger) Log(lvl LogLevel, message string) {
	l.logInternal(lvl, message, 2)
}

// logInternal ...
func (l *Logger) logInternal(lvl LogLevel, message string, pos int) {
	//var formatString string = "#%d %s [%s] %s:%d ▶ %.3s %s"
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
