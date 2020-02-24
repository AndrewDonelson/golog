// Package golog Simple flexible go logging
// This file contains all the code for the main logger
package golog

// Import packages
import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
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

	// FmtDefault is the default log format (QA)
	FmtDefault = "[%.16[3]s] #%[1]d %.19[2]s %.4[7]s %[8]s"
	// FmtProductionLog is the built-in production log format
	FmtProductionLog = "[%.16[3]s] %.19[2]s %.4[7]s - %[8]s"
	// FmtProductionJSON is the built-in production json format
	FmtProductionJSON = "{\"%.16[3]s\",\"%[5]s\",\"%[6]d\",\"%[4]s\",\"%[1]d\",\"%.19[2]s\",\"%[7]s\",\"%[8]s\"}"
	// FmtDevelopmentLog is the built-in development log format
	FmtDevelopmentLog = "[%.16[3]s] %.4[7]s %.19[2]s %[5]s#%[6]d-%[4]s : %[8]s"

	// Error, Fatal, Fatal Format
	//defErrorLogFmt = "\n%.8[7]s\nin %.16[3]s->%[4]s() file %[5]s on line %[6]d\n%[8]s\n"
)

var (
	// Log is set y the init function to be a default thelogger
	Log *Logger

	// Map for the various codes of colors
	colors map[LogLevel]string

	// Map from format's placeholders to printf verbs
	phfs map[string]string

	// Contains color strings for stdout
	logNo uint64

	defFmt = FmtDefault

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

// Log Level. Panic is not included as a level.
const (
	RawLevel     = iota + 1 // None
	ErrorLevel              // Red 		31 - Fatal  & Error Levels are same
	WarningLevel            // Yellow 	33
	SuccessLevel            // Green 	32
	NoticeLevel             // Cyan 	36
	InfoLevel               // White 	37
	DebugLevel              // Blue 	34
	TraceLevel              // Magneta	35
)

// Logger class that is an interface to user to log messages, Module is the module for which we are testing
// worker is variable of Worker class that is used in bottom layers to log the message
type Logger struct {
	Options Options
	started time.Time // Set once on initialization
	timer   time.Time // reset on each call to timeElapsed()
	worker  *Worker
}

func init() {
	Log = NewLogger(nil)
}

// NewLogger creates and returns new logger for the given model & environment
// module is the specific module for which we are logging
// environment overrides detected environment (if -1)
// color defines whether the output is to be colored or not, out is instance of type io.Writer defaults
// to os.Stderr
func NewLogger(opts *Options) *Logger {
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
	l := &Logger{worker: newWorker}
	l.Options = *opts
	l.init()
	return l
}

// init is called by NewLogger to detect running conditions and set all defaults
func (l *Logger) init() {
	// Set Testing flag to TRUE if testing detected
	l.Options.Testing = (flag.Lookup("test.v") != nil)

	l.timeReset()
	l.started = l.timer
	initColors()
	initFormatPlaceholders()
}

func (l *Logger) timeReset() {
	l.timer = time.Now()
}

func (l *Logger) timeElapsed(start time.Time) time.Duration {
	return time.Since(start)
}

func (l *Logger) timeLog(name string) {
	l.logInternal(InfoLevel, 2, fmt.Sprintf("%s took %v", name, l.timeElapsed(l.timer)))
}

// logInternal ...
func (l *Logger) logInternal(lvl LogLevel, pos int, a ...interface{}) {
	_, filename, line, _ := runtime.Caller(pos)
	msg := fmt.Sprintf("%v", a...)
	filename = path.Base(filename)
	info := &Info{
		ID:       atomic.AddUint64(&logNo, 1),
		Time:     time.Now().Format(l.worker.timeFormat),
		Module:   l.Options.Module,
		Level:    lvl,
		Message:  msg,
		Filename: filename,
		Line:     line,
		Duration: l.timeElapsed(l.timer),
		//format:   formatString,
	}
	l.worker.Log(lvl, 2, info)
}

func (l *Logger) traceInternal(pos int, a ...interface{}) {
	function, file, line := getCaller(pos)
	msg := fmt.Sprintf("%v", a...)
	file = path.Base(file)
	info := &Info{
		ID:       atomic.AddUint64(&logNo, 1),
		Time:     time.Now().Format(l.worker.timeFormat),
		Module:   l.Options.Module,
		Level:    TraceLevel,
		Message:  msg,
		Filename: file,
		Line:     line,
		Function: function,
		Duration: l.timeElapsed(l.timer),
		//format:   formatString,
	}
	l.worker.Log(info.Level, 2, info)
}

// SetModuleName sets the name of the module being logged
func (l *Logger) SetModuleName(name string) {
	l.Options.Module = name
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
func (l *Logger) SetEnvironment(env Environment) {
	l.Options.Environment = env
	l.worker.SetEnvironment(env)
}

// SetEnvironmentFromString is used to manually set the log environment to either development, testing or production
func (l *Logger) SetEnvironmentFromString(env string) {
	env = strings.ToLower(env)
	switch env {
	case "dev":
		l.SetEnvironment(EnvDevelopment)
	case "qa":
		l.SetEnvironment(EnvQuality)
	default:
		l.SetEnvironment(EnvProduction)
	}
}

// SetOutput is used to manually set the output to send log data
func (l *Logger) SetOutput(out io.Writer) {
	l.Options.Out = out
	l.worker.SetOutput(out)
}

// SetColor is used to manually set the color mode
func (l *Logger) SetColor(c ColorMode) {
	l.Options.UseColor = c
	l.worker.color = c
}

// UseJSONForProduction forces using JSON instead of log for production
func (l *Logger) UseJSONForProduction() {
	l.worker.UseJSONForProduction()
}

// Log The log command is the function available to user to log message,
// lvl specifies the degree of the message the user wants to log, message
// is the info user wants to log
func (l *Logger) Log(lvl LogLevel, a ...interface{}) {
	l.logInternal(lvl, 2, a...)
}

// PrettyPrint is used to display any type nicely in the log output
func (l *Logger) PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("\n%s\n", string(b))
}

// Trace is a basic timing function that will log InfoLevel duration of name
func (l *Logger) Trace(name, file string, line int) {
	l.timeReset()
	defer l.timeLog(name)
}

// Panic is just like func l.Fatal except that it is followed by a call to panic
func (l *Logger) Panic(a ...interface{}) {
	msg := fmt.Sprintf("%v", a...)
	l.logInternal(ErrorLevel, 2, a...)
	if l.Options.Testing {
		return
	}
	panic(msg)
}

// PanicE logs a error at Fatallevel
func (l *Logger) PanicE(err error) {
	l.Panic(err.Error())
}

// Panicf is just like func l.FatalF except that it is followed by a call to panic
func (l *Logger) Panicf(format string, a ...interface{}) {
	l.Panic(fmt.Sprintf(format, a...))
}

// Fatal is just like func l.Fatal logger except that it is followed by exit to program
func (l *Logger) Fatal(a ...interface{}) {
	l.logInternal(ErrorLevel, 2, a...)
	if l.Options.Testing {
		return
	}
	os.Exit(0)
}

// FatalE logs a error at Fatallevel
func (l *Logger) FatalE(err error) {
	l.Fatal(err.Error())
}

// Fatalf is just like func l.FatalF logger except that it is followed by exit to program
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.Fatal(fmt.Sprintf(format, a...))
}

// Error logs a customer message at Error level
func (l *Logger) Error(a ...interface{}) {
	l.logInternal(ErrorLevel, 2, a...)
}

// ErrorE logs a error at Error level
func (l *Logger) ErrorE(err error) {
	l.logInternal(ErrorLevel, 2, err.Error())
}

// Errorf logs a message at Error level using the same syntax and options as fmt.Printf
func (l *Logger) Errorf(format string, a ...interface{}) {
	l.logInternal(ErrorLevel, 2, fmt.Sprintf(format, a...))
}

// Warning logs a message at Warning level
func (l *Logger) Warning(a ...interface{}) {
	l.logInternal(WarningLevel, 2, a...)
}

// WarningE logs a error at Warning level
func (l *Logger) WarningE(err error) {
	l.Warning(err.Error())
}

// Warningf logs a message at Warning level using the same syntax and options as fmt.Printf
func (l *Logger) Warningf(format string, a ...interface{}) {
	l.logInternal(WarningLevel, 2, fmt.Sprintf(format, a...))
}

// Success logs a message at Success level
func (l *Logger) Success(a ...interface{}) {
	l.logInternal(SuccessLevel, 2, a...)
}

// Successf logs a message at Success level using the same syntax and options as fmt.Printf
func (l *Logger) Successf(format string, a ...interface{}) {
	l.logInternal(SuccessLevel, 2, fmt.Sprintf(format, a...))
}

// Notice logs a message at Notice level
func (l *Logger) Notice(a ...interface{}) {
	l.logInternal(NoticeLevel, 2, a...)
}

// Noticef logs a message at Notice level using the same syntax and options as fmt.Printf
func (l *Logger) Noticef(format string, a ...interface{}) {
	l.logInternal(NoticeLevel, 2, fmt.Sprintf(format, a...))
}

// Info logs a message at Info level
func (l *Logger) Info(a ...interface{}) {
	l.logInternal(InfoLevel, 2, a...)
}

// Infof logs a message at Info level using the same syntax and options as fmt.Printf
func (l *Logger) Infof(format string, a ...interface{}) {
	l.logInternal(InfoLevel, 2, fmt.Sprintf(format, a...))
}

// Debug logs a message at Debug level
func (l *Logger) Debug(a ...interface{}) {
	l.logInternal(DebugLevel, 2, a...)
}

// DebugE logs a error at Debug level
func (l *Logger) DebugE(err error) {
	l.Debug(err.Error())
}

// Debugf logs a message at Debug level using the same syntax and options as fmt.Printf
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.logInternal(DebugLevel, 2, fmt.Sprintf(format, a...))
}

// HandlerLog Traces & logs a message at Debug level for a REST handler
func (l *Logger) HandlerLog(w http.ResponseWriter, r *http.Request) {
	l.timeReset()
	defer l.traceInternal(4, fmt.Sprintf("%s %s %v", r.Method, r.RequestURI, l.timeElapsed(l.timer)))
}

// HandlerLogf logs a message at Debug level using the same syntax and options as fmt.Printf
func (l *Logger) HandlerLogf(w http.ResponseWriter, r *http.Request, format string, a ...interface{}) {
	l.timeReset()
	defer l.logInternal(DebugLevel, 2, fmt.Sprintf(format, a...))
}

// Print logs a message at directly with no level (RAW)
func (l *Logger) Print(a ...interface{}) {
	l.logInternal(RawLevel, 2, a...)
}

// Printf logs a message at Info level using the same syntax and options as fmt.Printf
func (l *Logger) Printf(format string, a ...interface{}) {
	l.logInternal(RawLevel, 2, fmt.Sprintf(format, a...))
}

// StackAsError Prints this goroutine's execution stack as an error with an optional message at the begining
func (l *Logger) StackAsError(message string) {
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	l.logInternal(ErrorLevel, 2, message+Stack())
}

// StackAsFatal Prints this goroutine's execution stack as fatal with an optional message at the begining
func (l *Logger) StackAsFatal(message string) {
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	l.logInternal(ErrorLevel, 2, message+Stack())
}

// Stack Returns a string with the execution stack for this goroutine
func Stack() string {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	return string(buf)
}
