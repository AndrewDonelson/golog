package golog

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func BenchmarkLoggerLog(b *testing.B) {
	b.StopTimer()
	log, err := New("test", 1)
	if err != nil {
		panic(err)
	}

	var tests = []struct {
		level   LogLevel
		message string
	}{
		{
			CriticalLevel,
			"Critical Logging",
		},
		{
			ErrorLevel,
			"Error logging",
		},
		{
			SuccessLevel,
			"Success logging",
		},
		{
			WarningLevel,
			"Warning logging",
		},
		{
			NoticeLevel,
			"Notice Logging",
		},
		{
			InfoLevel,
			"Info Logging",
		},
		{
			DebugLevel,
			"Debug logging",
		},
	}

	b.StartTimer()
	for _, test := range tests {
		for n := 0; n <= b.N; n++ {
			log.Log(test.level, test.message)
		}
	}
}

func BenchmarkLoggerNew(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		log, err := New("test", 1)
		if err != nil && log == nil {
			panic(err)
		}
	}
}

func BenchmarkLoggerNewLogger(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		log, err := NewLogger("bench-production", 1, nil)
		if err != nil && log == nil {
			panic(err)
		}
	}
}

func TestParseFormat(t *testing.T) {
	msgFmt, tmeFmt := parseFormat("foobar")
	want := fmt.Sprintf("%s, %s", defFmt, defTimeFmt)
	have := fmt.Sprintf("%s, %s", msgFmt, tmeFmt)
	if have != want {
		t.Errorf("\nWant: %s\nHave: %s", want, have)
	}

	msgFmt, tmeFmt = parseFormat("{%.10s} - Foobar")
	want = "{%%.10s} - Foobar, 2006-01-02 15:04:05"
	have = fmt.Sprintf("%s, %s", msgFmt, tmeFmt)
	if have != want {
		t.Errorf("\nWant: %s\nHave: %s", want, have)
	}
	
} 

func TestLoggerNew(t *testing.T) {
	log, err := New("test", 1)
	if err != nil {
		panic(err)
	}
	if log.Module != "test" {
		t.Errorf("Unexpected module: %s", log.Module)
	}
	log.SetFunction("TestLoggerNew")
	log.SetEnvironment(2)
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer

	// Test for no user defined out
	log, err := NewLogger("test", 1, nil)
	if err != nil {
		t.Error("Unexpected error. Wanted valid logger")
	}
	log.SetLogLevel(DebugLevel)

	// Test for module name less than 4 characters in length
	log, err = NewLogger("tst", 1, nil)
	if err == nil || log != nil {
		t.Error("Expected an error")
	} 

	// test with standard out
	log, err = NewLogger("test", 1, &buf)
	if err != nil {
		panic(err)
	}
	if log.Module != "test" {
		t.Errorf("Unexpected module: %s", log.Module)
	}
	log.SetFunction("TestLoggerNew")
	log.SetEnvironment(2)
	log.Log(WarningLevel,"This is only a warning")

	//log.Fatal("This is a fatal message")
	//log.Fatalf("This is %d %s message", 1, "fatal")
	log.Errorf("This is %d %s message", 1, "error")
	log.Successf("This is %d %s message", 1, "success")
	log.Warningf("This is %d %s message", 1, "warning")
	log.Noticef("This is %d %s message", 1, "notice")
	log.Infof("This is %d %s message", 1, "info")
	log.Debugf("This is %d %s message", 1, "debug")

	log.StackAsError("Stack as Error")
	log.StackAsCritical("Stack as Critical")
}

func TestColorString(t *testing.T) {
	colorCode := colorString(40)
	if colorCode != "\033[40m" {
		t.Errorf("Unexpected string: %s", colorCode)
	}
}

// CriticalLevel LogLevel = iota + 1 // Magneta 	35
// ErrorLevel                        // Red 		31
// SuccessLevel                      // Green 		32
// WarningLevel                      // Yellow 		33
// NoticeLevel                       // Cyan 		36
// InfoLevel                         // White 		37
// DebugLevel                        // Blue 		34

func TestInitColors(t *testing.T) {
	//initColors()
	var tests = []struct {
		level       LogLevel
		color       int
		colorString string
	}{
		{
			CriticalLevel,
			Magenta,
			"\033[35m",
		},
		{
			ErrorLevel,
			Red,
			"\033[31m",
		},
		{
			SuccessLevel,
			Green,
			"\033[32m",
		},
		{
			WarningLevel,
			Yellow,
			"\033[33m",
		},
		{
			NoticeLevel,
			Cyan,
			"\033[36m",
		},
		{
			InfoLevel,
			White,
			"\033[37m",
		},
		{
			DebugLevel,
			Blue,
			"\033[34m",
		},
	}

	for _, test := range tests {
		if colors[test.level] != test.colorString {
			t.Errorf("Unexpected color string %d", test.color)
		}
	}
}

func TestNewWorker(t *testing.T) {
	var worker = NewWorker("", 0, 1, os.Stderr)
	if worker.Minion == nil {
		t.Errorf("Minion was not established")
	}
}

func BenchmarkNewWorker(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		worker := NewWorker("", 0, 1, os.Stderr)
		if worker == nil {
			panic("Failed to initiate worker")
		}
	}
}

func TestLogger_SetFormat(t *testing.T) {
	var buf bytes.Buffer
	log, err := New("pkgname", 0, &buf)
	if err != nil || log == nil {
		panic(err)
	}

	log.SetLogLevel(DebugLevel)
	log.Debug("Test")
	log.SetLogLevel(InfoLevel)

	//want := time.Now().Format("2006-01-02 15:04:05")
	want := fmt.Sprintf("#10 %s golog_test.go:237 ▶ DEB Test\n", time.Now().Format("2006-01-02 15:04:05"))
	have := buf.String()
	if have != want {
		t.Errorf("\nWant: %sHave: %s", want, have)
	}
	format :=
		"text123 %{id} " + // text and digits before id
			"!@#$% %{time:Monday, 2006 Jan 01, 15:04:05} " + // symbols before time with spec format
			"a{b %{module} " + // brace with text that should be just text before verb
			"a}b %{filename} " + // brace with text that should be just text before verb
			"%% %{file} " + // percent symbols before verb
			"%{%{line} " + // percent symbol with brace before verb w/o space
			"%{nonex_verb} %{lvl} " + // nonexistent verb berfore real verb
			"%{incorr_verb %{level} " + // incorrect verb before real verb
			"%{} [%{message}]" // empty verb before message in sq brackets
	buf.Reset()
	log.SetFormat(format)
	log.Error("This is Error!")
	now := time.Now()
	want = fmt.Sprintf(
		"text123 11 "+
			"!@#$%% %s "+
			"a{b pkgname "+
			"a}b golog_test.go "+
			"%%%% golog_test.go "+ // it's printf, escaping %, don't forget
			"%%{258 "+
			" ERR "+
			"%%{incorr_verb ERROR "+
			" [This is Error!]\n",
		now.Format("Monday, 2006 Jan 01, 15:04:05"),
	)
	have = buf.String()
	if want != have {
		t.Errorf("\nWant: %sHave: %s", want, have)
		wantLen := len(want)
		haveLen := len(have)
		min := int(math.Min(float64(wantLen), float64(haveLen)))
		if wantLen != haveLen {
			t.Errorf("Diff lens: Want: %d, Have: %d.\n", wantLen, haveLen)
		}
		for i := 0; i < min; i++ {
			if want[i] != have[i] {
				t.Errorf("Differents starts at %d pos (\"%c\" != \"%c\")\n",
					i, want[i], have[i])
				break
			}
		}
	}
}

func TestSetDefaultFormat(t *testing.T) {
	SetDefaultFormat("%{module} %{lvl} %{message}")
	var buf bytes.Buffer
	log, err := New("pkgname", 0, &buf)
	if err != nil || log == nil {
		panic(err)
	}
	log.Criticalf("Test %d", 123)
	want := "pkgname CRI Test 123\n"
	have := buf.String()
	if want != have {
		t.Errorf("\nWant: %sHave: %s", want, have)
	}
}

func TestLogLevel(t *testing.T) {

	var tests = []struct {
		level   LogLevel
		message string
	}{
		{
			CriticalLevel,
			"Critical Logging",
		},
		{
			ErrorLevel,
			"Error logging",
		},
		{
			SuccessLevel,
			"Success logging",
		},
		{
			WarningLevel,
			"Warning logging",
		},
		{
			NoticeLevel,
			"Notice Logging",
		},
		{
			InfoLevel,
			"Info Logging",
		},
		{
			DebugLevel,
			"Debug logging",
		},
	}

	var buf bytes.Buffer
	log, err := New("pkgname", 0, &buf)
	if err != nil {
		panic(err)
	}

	for i, test := range tests {
		log.SetLogLevel(test.level)

		log.Critical("Log Critical")
		log.Error("Log Error")
		log.Success("Log Success")
		log.Warning("Log Warning")
		log.Notice("Log Notice")
		log.Info("Log Info")
		log.Debug("Log Debug")

		// Count output lines from logger
		count := strings.Count(buf.String(), "\n")
		if i+1 != count {
			t.Error()
		}
		buf.Reset()
	}
}
