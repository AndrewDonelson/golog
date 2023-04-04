package golog

import (
	"bytes"
	"errors"
	"strings"
)

// ErrInvalidLevel is returned if the severity level is invalid.
var ErrInvalidLevel = errors.New("invalid level")

// Level of severity.
type Level int

// Log levels.
const (
	InvalidLevel = iota + 1 // None
	RawLevel                // None
	ErrorLevel              // Red 		31 - Fatal  & Error Levels are same
	TraceLevel              // Magneta	35
	WarningLevel            // Yellow 	33
	SuccessLevel            // Green 	32
	NoticeLevel             // Cyan 	36
	InfoLevel               // White 	37
	DebugLevel              // Blue 	34
)

var levelNames = [...]string{
	RawLevel:     "raw",
	TraceLevel:   "trace",
	WarningLevel: "warning",
	SuccessLevel: "success",
	NoticeLevel:  "notice",
	InfoLevel:    "info",
	DebugLevel:   "debug",
	ErrorLevel:   "error",
}

var levelStrings = map[string]Level{
	"raw":     RawLevel,
	"trace":   TraceLevel,
	"warning": WarningLevel,
	"success": SuccessLevel,
	"notice":  NoticeLevel,
	"info":    InfoLevel,
	"debug":   DebugLevel,
	"error":   ErrorLevel,
}

// String implementation.
func (l Level) String() string {
	return levelNames[l]
}

// MarshalJSON implementation.
func (l Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.String() + `"`), nil
}

// UnmarshalJSON implementation.
func (l *Level) UnmarshalJSON(b []byte) error {
	v, err := ParseLevel(string(bytes.Trim(b, `"`)))
	if err != nil {
		return err
	}

	*l = v
	return nil
}

// ParseLevel parses level string.
func ParseLevel(s string) (Level, error) {
	l, ok := levelStrings[strings.ToLower(s)]
	if !ok {
		return InvalidLevel, ErrInvalidLevel
	}

	return l, nil
}

// MustParseLevel parses level string or panics.
func MustParseLevel(s string) Level {
	l, err := ParseLevel(s)
	if err != nil {
		panic("invalid log level")
	}

	return l
}
