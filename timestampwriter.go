package horenso

import (
	"bytes"
	"io"
	"sync"
	"time"
)

type timestampWriter struct {
	writer    io.Writer
	midOfLine bool
	m         *sync.Mutex
}

func newTimestampWriter(w io.Writer) *timestampWriter {
	return &timestampWriter{
		writer: w,
		m:      &sync.Mutex{},
	}
}

func (l *timestampWriter) Write(buf []byte) (int, error) {
	l.m.Lock()
	defer l.m.Unlock()

	var bb bytes.Buffer
	for i, chr := range buf {
		if !l.midOfLine {
			bb.Write(timestampBytes())
			l.midOfLine = true
		}
		if chr == '\n' {
			l.midOfLine = false
		}
		bb.Write(buf[i : i+1])
	}
	l.writer.Write(bb.Bytes())
	return len(buf), nil
}

const layout = "[2006-01-02 15:04:05.999999] "

func timestampBytes() []byte {
	t := time.Now()
	b := make([]byte, 0, len(layout))
	return t.AppendFormat(b, layout)
}
