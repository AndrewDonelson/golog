package golog

// memoryBackend should be used only in unit tests
var _ Backend = (*memoryBackend)(nil)

type memoryBackend struct {
	records []*Record
	msgs    []string
}

func (b *memoryBackend) Log(level Level, calldepth int, record *Record) error {
	b.records = append(b.records, record)
	return nil
}

func (b *memoryBackend) LogStr(level Level, calldepth int, msg string) error {
	b.msgs = append(b.msgs, msg)
	return nil
}

func (b *memoryBackend) GetFormatter() Formatter { return nil }

func newMemoryBackend() Backend {
	return &memoryBackend{
		records: make([]*Record, 0),
		msgs:    make([]string, 0),
	}
}
