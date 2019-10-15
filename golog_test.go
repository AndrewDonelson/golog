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
	log, err := NewLogger(nil)
	if err != nil {
		panic(err)
	}
	log.Options.Module = "BenchLog"

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
		log, err := NewLogger(nil)
		if err != nil && log == nil {
			panic(err)
		}
	}
}

func BenchmarkLoggerNewLogger(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		log, err := NewLogger(nil)
		if err != nil && log == nil {
			panic(err)
		}
		log.Options.Module = "BenchNewLogger"
		log.SetEnvironment(0)
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
	log, err := NewLogger(NewDefaultOptions())
	if err != nil {
		t.Error(err)
		return
	}

	if log.Module != "unknown" {
		t.Errorf("Unexpected module: %s", log.Module)
	}

	log.SetFunction("TestLoggerNew")
	log.SetEnvironment(2)
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer

	// Test for no user defined out
	log, err := NewLogger(NewDefaultOptions())
	if err != nil {
		t.Error("Unexpected error. Wanted valid logger")
	}
	log.SetLogLevel(DebugLevel)

	// test with standard out
	log, err = NewLogger(&Options{
		Module: "test",
		Out:    &buf,
	})
	if err != nil {
		t.Error(err)
		return
	}
	if log.Module != "test" {
		t.Errorf("Unexpected module: %s", log.Module)
	}
	log.SetFunction("TestLoggerNew")
	log.SetEnvironment(2)
	log.Log(WarningLevel, "This is only a warning")

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

func TestNewWorker(t *testing.T) {
	var worker = NewWorker("", 0, true, os.Stderr)
	if worker.Minion == nil {
		t.Errorf("Minion was not established")
	}
}

func BenchmarkNewWorker(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		worker := NewWorker("", 0, true, os.Stderr)
		if worker == nil {
			panic("Failed to initiate worker")
		}
	}
}

func TestLogger_SetFormat(t *testing.T) {
	var buf bytes.Buffer
	log, err := NewLogger(&Options{
		Module:      "pkgname",
		Out:         &buf,
		Environment: 0,
		UseColor:    false,
	})
	if err != nil || log == nil {
		t.Error(err)
		return
	}

	log.SetLogLevel(DebugLevel)
	log.Debug("Test")
	log.SetLogLevel(InfoLevel)

	want := fmt.Sprintf("pkgname %s DEB ▶ Test\n", time.Now().Format("2006-01-02 15:04:05"))
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
			"%%{261 "+
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
	var buf bytes.Buffer

	log, err := NewLogger(&Options{
		Module:      "pkgname",
		Out:         &buf,
		Environment: 0,
		UseColor:    false,
	})
	if err != nil || log == nil {
		t.Error(err)
		return
	}

	SetDefaultFormat(defFmt)

	now := time.Now()
	log.Criticalf("Test %d", 123)
	want := fmt.Sprintf("pkgname %s CRI ▶ Test 123\n", now.Format("2006-01-02 15:04:05"))
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

	log, err := NewLogger(&Options{
		Module:      "pkgname",
		Out:         &buf,
		Environment: 0,
		UseColor:    false,
	})
	if err != nil || log == nil {
		t.Error(err)
		return
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
