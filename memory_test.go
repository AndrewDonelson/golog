package golog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getRecord() *Record {
	// Make a test run to make sure we can format it correctly.
	t, err := time.Parse(time.RFC3339, "2010-02-04T21:00:57-08:00")
	if err != nil {
		panic(err)
	}
	testFmt := "hello %s"
	r := &Record{
		ID:     12345,
		Time:   t,
		Module: "logger",
		Args:   []interface{}{"go"},
		fmt:    &testFmt,
	}

	return r
}
func TestMemory(t *testing.T) {
	rec := getRecord()
	backend := newMemoryBackend()

	assert.NotNil(t, backend)
	assert.NoError(t, backend.Log(DEBUG, 1, rec))
	assert.NoError(t, backend.LogStr(WARNING, 1, "Hello World!"))
	assert.Nil(t, backend.GetFormatter())
}
