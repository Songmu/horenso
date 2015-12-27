package horenso

import (
	"bytes"
	"regexp"
	"testing"
)

func TestTimeStampWriter(t *testing.T) {
	var bb bytes.Buffer
	l := newTimestampWriter(&bb)
	l.Write([]byte{'c'})
	l.Write([]byte{'d'})
	l.Write([]byte{'\n', 'e', 'f'})

	timeRegStr := `\[[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{1,6}\] `
	reg := regexp.MustCompile(`\A` + timeRegStr + "cd\n" + timeRegStr + `ef\z`)
	if !reg.Match(bb.Bytes()) {
		t.Errorf("something went wrong. output: %s", bb.String())
	}
}
