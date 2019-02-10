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

func timestampBytes() []byte {
	return formatTimestamp(time.Now())
}

const layout = "[2006-01-02 15:04:05.999999"

func formatTimestamp(t time.Time) []byte {
	b := make([]byte, 0, len(layout)+2) // layout + "] "
	b = t.AppendFormat(b, layout)
	// 20 == len("[2006-01-02 15:04:05")
	if len(b) == 20 {
		b = append(b, '.')
	}
	for l := len(b); l < len(layout); l++ {
		b = append(b, '0')
	}
	b = append(b, ']', ' ')
	return b
}
