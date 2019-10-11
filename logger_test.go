package logging

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	module           = "module"
	contextSeparator = ": "
	bufferCapacity   = 32
	format           = "%{message}"

	msgTemplate = "%s"
	msgDebug    = "debug"
	msgInfo     = "info"
	msgNotice   = "notice"
	msgWarning  = "warning"
	msgError    = "error"
	msgCritical = "critical"
)

func loggerWithMemoryBackend() *Logger {
	l := MustGetLogger(module, contextSeparator, bufferCapacity)
	l.SetBackend(newMemoryBackend())
	l.SetFormatter(MustStringFormatter(format))
	return l
}

func TestLogger_Debug_SimpleMethods(t *testing.T) {
	l := loggerWithMemoryBackend()
	l.SetLevel(DEBUG)

	l.Debug(msgDebug)
	l.Info(msgInfo)
	l.Notice(msgNotice)
	l.Warning(msgWarning)
	l.Error(msgError)
	l.Critical(msgCritical)

	// Check records that were received by backend
	backend := l.backend.(*memoryBackend)
	assert.Len(t, backend.records, 6)
	assert.Len(t, backend.msgs, 0)
	assert.Equal(t, msgDebug, backend.records[0].Args[0])
	assert.Equal(t, msgInfo, backend.records[1].Args[0])
	assert.Equal(t, msgNotice, backend.records[2].Args[0])
	assert.Equal(t, msgWarning, backend.records[3].Args[0])
	assert.Equal(t, msgError, backend.records[4].Args[0])
	assert.Equal(t, msgCritical, backend.records[5].Args[0])

	// Ring buffer should be empty, because we're in debug
	assert.Nil(t, l.records.Dequeue())
}

func TestLogger_Debug_FormattingMethods(t *testing.T) {
	l := loggerWithMemoryBackend()
	l.SetLevel(DEBUG)

	l.Debugf(msgTemplate, msgDebug)
	l.Infof(msgTemplate, msgInfo)
	l.Noticef(msgTemplate, msgNotice)
	l.Warningf(msgTemplate, msgWarning)
	l.Errorf(msgTemplate, msgError)
	l.Criticalf(msgTemplate, msgCritical)

	// Check records that were received by backend
	backend := l.backend.(*memoryBackend)
	assert.Len(t, backend.records, 6)
	assert.Len(t, backend.msgs, 0)
	assert.Equal(t, msgDebug, backend.records[0].Args[0])
	assert.Equal(t, msgInfo, backend.records[1].Args[0])
	assert.Equal(t, msgNotice, backend.records[2].Args[0])
	assert.Equal(t, msgWarning, backend.records[3].Args[0])
	assert.Equal(t, msgError, backend.records[4].Args[0])
	assert.Equal(t, msgCritical, backend.records[5].Args[0])

	// Ring buffer should be empty, because we're in debug
	assert.Nil(t, l.records.Dequeue())
}

func TestLogger_Dumping(t *testing.T) {
	l := loggerWithMemoryBackend()
	l.SetLevel(ERROR)
	l.SetDumpBehavior(true, ERROR, "")

	// Send several messages that have level less than loggers level
	l.Debug(msgDebug)
	l.Info(msgInfo)
	l.Notice(msgNotice)
	l.Warningf(msgTemplate, msgWarning)

	// They must not exist in backend yet
	backend := l.backend.(*memoryBackend)
	assert.Len(t, backend.records, 0)

	// Send message with triggering level
	l.Error(msgError)

	// Now we suppose to have 1 error record and 4 previous messages in backend
	assert.Len(t, backend.records, 1)
	assert.Len(t, backend.msgs, 4)
	assert.Equal(t, msgError, backend.records[0].Args[0])
	assert.Equal(t, msgDebug, backend.msgs[0])
	assert.Equal(t, msgInfo, backend.msgs[1])
	assert.Equal(t, msgNotice, backend.msgs[2])
	assert.Equal(t, msgWarning, backend.msgs[3])

	// Ring buffer is empty now
	assert.Nil(t, l.records.Dequeue())
}

func TestLogger_RingBufferReusage(t *testing.T) {
	l := MustGetLogger(module, contextSeparator, 4)
	l.SetBackend(newMemoryBackend())
	l.SetFormatter(MustStringFormatter(format))
	l.SetLevel(ERROR)
	l.SetDumpBehavior(true, ERROR, "")

	// Write 8 low-level messages
	for i := 0; i < 8; i++ {
		l.Debug(i)
	}

	// Flush ring buffer into backend
	l.Error(msgError)

	backend := l.backend.(*memoryBackend)
	assert.Len(t, backend.records, 1)
	assert.Len(t, backend.msgs, 4)

	// Only last 4 low-level message should appear in backend
	for i := 4; i < 8; i++ {
		assert.Equal(t, fmt.Sprint(i), backend.msgs[i-4])
	}
}
