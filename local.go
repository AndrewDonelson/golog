// Package golog Simple flexible go logging
// This file contains all un-exported (local) functions
package golog

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func detectEnvironment(testing bool) Environment {
	if testing && flag.Lookup("test.v") != nil {
		return EnvTesting
	}

	be := os.Getenv("BUILD_ENV")
	if be == "dev" {
		return EnvDevelopment
	} else if be == "qa" {
		return EnvQuality
	}

	return EnvProduction
}

// Analyze and represent format string as printf format string and time format
func parseFormat(format string) (msgfmt, timefmt string) {
	if len(format) < 10 /* (len of "%{message} */ {
		return defFmt, defTimeFmt
	}
	timefmt = defTimeFmt
	idx := strings.IndexRune(format, '%')
	for idx != -1 {
		msgfmt += format[:idx]
		format = format[idx:]
		if len(format) > 2 {
			if format[1] == '{' {
				// end of curr verb pos
				if jdx := strings.IndexRune(format, '}'); jdx != -1 {
					// next verb pos
					idx = strings.Index(format[1:], "%{")
					// incorrect verb found ("...%{wefwef ...") but after
					// this, new verb (maybe) exists ("...%{inv %{verb}...")
					if idx != -1 && idx < jdx {
						msgfmt += "%%"
						format = format[1:]
						continue
					}
					// get verb and arg
					verb, arg := ph2verb(format[:jdx+1])
					msgfmt += verb
					// check if verb is time
					// here you can handle args for other verbs
					if verb == `%[2]s` && arg != "" /* %{time} */ {
						timefmt = arg
					}
					format = format[jdx+1:]
				} else {
					format = format[1:] // TODO: Hit with test
				}
			} else {
				msgfmt += "%%"
				format = format[1:]
			}
		}
		idx = strings.IndexRune(format, '%')
	}
	msgfmt += format
	return
}

// translate format placeholder to printf verb and some argument of placeholder
// (now used only as time format)
func ph2verb(ph string) (verb string, arg string) {
	n := len(ph)
	if n < 4 {
		return ``, ``
	}
	if ph[0] != '%' || ph[1] != '{' || ph[n-1] != '}' {
		return ``, `` // TODO: Hit with test
	}
	idx := strings.IndexRune(ph, ':')
	if idx == -1 {
		return phfs[ph], ``
	}
	verb = phfs[ph[:idx]+"}"]
	arg = ph[idx+1 : n-1]
	return
}

// colorString Returns a proper string to output for colored logging
func colorString(color int) string {
	return fmt.Sprintf("\033[%dm", int(color))
}

// initColors Initializes the map of colors
func initColors() {
	colors = map[LogLevel]string{
		RawLevel:      colorString(White),
		CriticalLevel: colorString(Magenta),
		ErrorLevel:    colorString(Red),
		SuccessLevel:  colorString(Green),
		WarningLevel:  colorString(Yellow),
		NoticeLevel:   colorString(Cyan),
		InfoLevel:     colorString(White),
		DebugLevel:    colorString(Blue),
	}
}

// initFormatPlaceholders Initializes the map of placeholders
// "%{id}, %{time}, %{module}, %{function}, %{filename}, %{file}, %{line}, %{level}, %{lvl}, %{message}"
func initFormatPlaceholders() {
	phfs = map[string]string{
		"%{id}":         "%[1]d",
		"%{time}":       "%[2]s",
		"%{module}":     "%[3]s",
		"%{function}":   "%[4]s",
		"%{filename}":   "%[5]s",
		"%{file}":       "%[5]s",
		"%{line}":       "%[6]d",
		"%{level}":      "%[7]s",
		"%{lvl}":        "%.3[7]s",
		"%{message}":    "%[8]s",
		"%{duration}":   "%[9]s",
		"%{method}":     "%[10]s",
		"%{statuscode}": "%[11]d",
		"%{route}":      "%[12]s",
	}
}

func getCaller(skipLevels int) (function, file string, line int) {
	fpcs := make([]uintptr, 1)
	// Skip `skipLevels` levels to get the caller
	n := runtime.Callers(skipLevels, fpcs)
	if n != 0 {
		caller := runtime.FuncForPC(fpcs[0] - 1)
		if caller != nil {
			file, line = caller.FileLine(fpcs[0] - 1)
			function = caller.Name()
		}
	}

	return
}
