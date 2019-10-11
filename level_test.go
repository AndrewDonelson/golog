package golog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevel_IntToStr(t *testing.T) {
	assert.Len(t, levelNames, 7)
	assert.Equal(t, CRITICAL.String(), "CRITICAL")
	assert.Equal(t, ERROR.String(), "ERROR")
	assert.Equal(t, WARNING.String(), "WARNING")
	assert.Equal(t, SUCCESS.String(), "SUCCESS")
	assert.Equal(t, NOTICE.String(), "NOTICE")
	assert.Equal(t, INFO.String(), "INFO")
	assert.Equal(t, DEBUG.String(), "DEBUG")
}

func TestLevel_StrToInt(t *testing.T) {
	var value Level
	assert.Len(t, levelNames, 7)

	value, _ = LogLevel("CRITICAL")
	assert.Equal(t, value, CRITICAL)

	value, _ = LogLevel("ERROR")
	assert.Equal(t, value, ERROR)

	value, _ = LogLevel("WARNING")
	assert.Equal(t, value, WARNING)

	value, _ = LogLevel("SUCCESS")
	assert.Equal(t, value, SUCCESS)

	value, _ = LogLevel("NOTICE")
	assert.Equal(t, value, NOTICE)

	value, _ = LogLevel("INFO")
	assert.Equal(t, value, INFO)

	value, _ = LogLevel("DEBUG")
	assert.Equal(t, value, DEBUG)
}
