package golog

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAdvancedFormat(t *testing.T) {
	var buf bytes.Buffer
	log, err := NewLogger(nil)
	if err != nil || log == nil {
		t.Error(err)
		return
	}
	log.SetModuleName("pkgname")
	log.SetOutput(&buf)
	log.SetEnvironment(EnvDevelopment)
	log.SetColor(ClrNotSet)

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
	log.SetFormat(format)
	log.Error("This is Error!")
	now := time.Now()
	want := fmt.Sprintf(
		"text123 1 "+ //SET TO 1 for running this test alone and SET TO 11 for running as package test
			"!@#$%% %s "+
			"a{b pkgname "+
			"a}b golog_test.go "+
			"%%%% golog_test.go "+ // it's printf, escaping %, don't forget
			"%%{38 "+
			" ERR "+
			"%%{incorr_verb ERROR "+
			" [This is Error!]\n",
		now.Format("Monday, 2006 Jan 01, 15:04:05"),
	)

	have := buf.String()
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
				t.Errorf("Differents starts at %d pos (\"%c\" != \"%c\")\n", i, want[i], have[i])
				break
			}
		}
	}

}

// func TestLogger_SetFormat(t *testing.T) {
// 	var buf bytes.Buffer
// 	log, err := NewLogger(&Options{
// 		Module: "pkgname",
// 		Out:    &buf,
// 	})
// 	if err != nil || log == nil {
// 		t.Error(err)
// 		return
// 	}

// 	log.SetLogLevel(DebugLevel)
// 	log.SetFunction("TestLogger_SetFormat")
// 	log.SetFormat(FmtDevelopmentLog)
// 	log.Debug("Test")
// 	//log.SetLogLevel(InfoLevel)

// 	want := fmt.Sprintf("[34m[pkgname] %s DEB - golog_test.go#86-TestLogger_SetFormat - Test[0m\n", time.Now().Format("2006-01-02 15:04:05"))
// 	have := buf.String()
// 	if have != want {
// 		t.Errorf("\nWant: %sHave: %s", want, have)
// 	}
// }
func TestBuildEnvironments(t *testing.T) {
	os.Setenv("BUILD_ENV", "dev")
	if detectEnvironment() != EnvDevelopment {
		t.Error("Failed to SetEnvironment to EnvDevelopment")
	}

	os.Setenv("BUILD_ENV", "qa")
	if detectEnvironment() != EnvQuality {
		t.Error("Failed to SetEnvironment to EnvQuality")
	}

	os.Setenv("BUILD_ENV", "prod")
	if detectEnvironment() != EnvProduction {
		t.Error("Failed to SetEnvironment to EnvProduction")
	}

	log, err := NewLogger(&Options{UseColor: ClrDisabled})
	if err != nil {
		t.Error(err)
		return
	}
	log.SetEnvironment(EnvAuto)
	log.SetEnvironment(EnvDevelopment)
	log.SetEnvironment(EnvQuality)
	log.SetEnvironment(EnvProduction)
}
func TestParseFormat(t *testing.T) {
	// We do this just to initialize the required code on the
	log, err := NewLogger(nil)
	if err != nil {
		t.Error(err)
		return
	}
	log.SetEnvironment(2)

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

	msgFmt, tmeFmt = parseFormat("%{id}, %{time}, %{module}, %{function}, %{filename}, %{file}, %{line}, %{level}, %{lvl}, %{message}")
	want = "%[1]d, %[2]s, %[3]s, %[4]s, %[5]s, %[5]s, %[6]d, %[7]s, %.3[7]s, %[8]s, 2006-01-02 15:04:05"
	have = fmt.Sprintf("%s, %s", msgFmt, tmeFmt)
	if have != want {
		t.Errorf("\nWant: %s\nHave: %s", want, have)
	}
}

func TestGlobalLogger(t *testing.T) {
	Log.Info("Testing default global logger")
}

func TestLoggerNew(t *testing.T) {
	log, err := NewLogger(NewDefaultOptions())
	if err != nil {
		t.Error(err)
		return
	}
	log.Trace("TestLoggerNew", "golang_test.go", 136)

	if log.Options.Module != "unknown" {
		t.Errorf("Unexpected module: %s", log.Options.Module)
	}

	log.SetFunction("TestLoggerNew")
	log.SetEnvironment(2)
	log.Log(CriticalLevel, "Testing 123")

	// Test for invalid output passed in
	log, err = NewLogger(&Options{Module: "BadOut", Out: nil})
	if err != nil || log == nil {
		t.Error(err)
		return
	}

	// Test for Module name to short < 4
	log, err = NewLogger(&Options{Module: "mod"})
	if err != nil || log == nil {
		t.Error(err)
		return
	}

	// Test for Module name to short < 4
	log, err = NewLogger(&Options{Module: "mod"})
	if err != nil || log == nil {
		t.Error(err)
		return
	}

	// Test for Module name to short < 4
	log, err = NewLogger(NewDefaultOptions())
	if err != nil || log == nil {
		t.Error(err)
		return
	}
	log.SetEnvironment(EnvProduction)
	log.UseJSONForProduction()

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
	log.SetOutput(&buf)

	log.SetModuleName("test")
	if log.Options.Module != "test" {
		t.Errorf("Unexpected module: %s", log.Options.Module)
	}

	log.SetFunction("TestLoggerNew")
	log.SetEnvironment(EnvDevelopment)

	log.Critical("This is a critial message")
	log.Fatal("This is a Fatal message")
	log.Panic("This is a Panic message")
	log.Error("This is a Error message")
	log.Success("This is a Success message")
	log.Warning("This is a Warning message")
	log.Notice("This is a Notice message")
	log.Info("This is a Info message")
	log.Debug("This is a Debug message")
	log.Print("This is a plain RAW Message")
	log.Trace("This is a trace message", "golog_test", 211)

	log.Criticalf("This is %d %s message", 1, "critical")
	log.Fatalf("This is %d %s message", 1, "fatal")
	log.Panicf("This is %d %s message", 1, "panic")
	log.Errorf("This is %d %s message", 1, "error")
	log.Successf("This is %d %s message", 1, "success")
	log.Warningf("This is %d %s message", 1, "warning")
	log.Noticef("This is %d %s message", 1, "notice")
	log.Infof("This is %d %s message", 1, "info")
	log.Debugf("This is %d %s message", 1, "debug")
	log.Printf("%s with %d args", "Message", 2)

	testErr := fmt.Errorf("Test Error")
	log.PanicE(testErr)
	log.CriticalE(testErr)
	log.FatalE(testErr)
	log.ErrorE(testErr)
	log.DebugE(testErr)
	log.WarningE(testErr)

	log.StackAsError("")
	log.StackAsCritical("")

	log.StackAsError("Stack as Error")
	log.StackAsCritical("Stack as Critical")
}

func TestNewloggerCustom(t *testing.T) {
	var buf bytes.Buffer
	log, err := NewLogger(NewCustomOptions(
		"modulename",
		EnvDevelopment,
		ClrAuto,
		true,
		&buf,
		FmtDefault,
		FmtDefault,
	))
	if err != nil || log == nil {
		t.Error("Unexpected error. Wanted valid logger")
	}

}

func TestPrettyPrint(t *testing.T) {
	var buf bytes.Buffer

	// Test for no user defined out
	log, err := NewLogger(NewDefaultOptions())
	if err != nil {
		t.Error("Unexpected error. Wanted valid logger")
	}
	log.SetLogLevel(DebugLevel)

	// test with standard out
	log.SetOutput(&buf)

	log.SetModuleName("pretty-print")
	if log.Options.Module != "pretty-print" {
		t.Errorf("Unexpected module: %s", log.Options.Module)
	}

	log.SetFunction("TestPrettyPrint")
	log.SetEnvironment(EnvDevelopment)

	log.Critical("Options", log.PrettyPrint(log.Options))
	log.Fatal("Options", log.PrettyPrint(log.Options))
	log.Panic("Options", log.PrettyPrint(log.Options))
	log.Error("Options", log.PrettyPrint(log.Options))
	log.Success("Options", log.PrettyPrint(log.Options))
	log.Warning("Options", log.PrettyPrint(log.Options))
	log.Notice("Options", log.PrettyPrint(log.Options))
	log.Info("Options", log.PrettyPrint(log.Options))
	log.Debug("Options", log.PrettyPrint(log.Options))
	log.Print("Options", log.PrettyPrint(log.Options))
}

func TestColorString(t *testing.T) {
	colorCode := colorString(40)
	if colorCode != "\033[40m" {
		t.Errorf("Unexpected string: %s", colorCode)
	}
}

func TestNewWorker(t *testing.T) {
	var worker = NewWorker("", 0, ClrNotSet, os.Stderr)
	if worker.Minion == nil {
		t.Errorf("Minion was not established")
	}
}

func BenchmarkNewWorker(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		worker := NewWorker("", 0, ClrNotSet, os.Stderr)
		if worker == nil {
			panic("Failed to initiate worker")
		}
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
			WarningLevel,
			"Warning logging",
		},
		{
			SuccessLevel,
			"Success logging",
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
		Out:         &buf,
		Environment: EnvDevelopment,
		UseColor:    ClrNotSet,
	})
	if err != nil || log == nil {
		t.Error(err)
		return
	}
	log.SetModuleName("pkgname")

	for i, test := range tests {
		log.SetLogLevel(test.level)

		log.Critical("Log Critical")
		log.Error("Log Error")
		log.Warning("Log Warning")
		log.Success("Log Success")
		log.Notice("Log Notice")
		log.Info("Log Info")
		log.Debug("Log Debug")

		// Count output lines from logger
		count := strings.Count(buf.String(), "\n")
		if i+1 != count {
			t.Errorf("Log events expected %d have %d", i+1, count)
		}
		buf.Reset()
	}
}

var golog *Logger

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	golog.HandlerLog(w, r)
	golog.HandlerLogf(w, r, "")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"message": "hello world"}`))
}
func TestHandlers(t *testing.T) {
	var (
		buf bytes.Buffer
	)

	golog, _ = NewLogger(&Options{Module: "test-handlers", Out: &buf})
	golog.SetEnvironment(EnvDevelopment)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeHTTP)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"message": "hello world"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

/*********************** BENCHMARKS *****************************/
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
