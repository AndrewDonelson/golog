// Package text implements a development-friendly textual handler.
package text

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/AndrewDonelson/golog"
)

// colors.
const (
	none    = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	magneta = 35
	cyan    = 36
	white   = 37
)

// Default handler outputting to stderr.
var Default = New(os.Stderr)

// start time.
var start = time.Now()

// Colors mapping.
var Colors = [...]int{
	golog.RawLevel:     white,
	golog.TraceLevel:   magneta,
	golog.WarningLevel: yellow,
	golog.SuccessLevel: green,
	golog.NoticeLevel:  blue,
	golog.InfoLevel:    white,
	golog.DebugLevel:   cyan,
	golog.ErrorLevel:   red,
}

// Strings mapping.
var Strings = [...]string{
	golog.RawLevel:     "raw",
	golog.TraceLevel:   "trace",
	golog.WarningLevel: "warning",
	golog.SuccessLevel: "success",
	golog.NoticeLevel:  "notice",
	golog.InfoLevel:    "info",
	golog.DebugLevel:   "debug",
	golog.ErrorLevel:   "error",
}

// Handler implementation.
type Handler struct {
	mu     sync.Mutex
	Writer io.Writer
}

// New handler.
func New(w io.Writer) *Handler {
	return &Handler{
		Writer: w,
	}
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(e *golog.Entry) error {
	color := Colors[e.Level]
	level := Strings[e.Level]
	names := e.Fields.Names()

	h.mu.Lock()
	defer h.mu.Unlock()

	ts := time.Since(start) / time.Second
	fmt.Fprintf(h.Writer, "\033[%dm%6s\033[0m[%04d] %-25s", color, level, ts, e.Message)

	for _, name := range names {
		fmt.Fprintf(h.Writer, " \033[%dm%s\033[0m=%v", color, name, e.Fields.Get(name))
	}

	fmt.Fprintln(h.Writer)

	return nil
}
