package horenso

import (
	"bytes"
	"regexp"
	"testing"
	"time"
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

func TestFormatTimestamp(t *testing.T) {
	testCases := []struct {
		name   string
		input  time.Time
		expect string
	}{
		{
			name:   "simple",
			input:  time.Date(2019, time.November, 4, 11, 12, 13, 123456000, time.UTC),
			expect: "[2019-11-04 11:12:13.123456] ",
		},
		{
			name:   "pad zero",
			input:  time.Date(2019, time.November, 9, 11, 12, 13, 123400000, time.UTC),
			expect: "[2019-11-09 11:12:13.123400] ",
		},
		{
			name:   "no millisec",
			input:  time.Date(2019, time.November, 3, 11, 12, 13, 0, time.UTC),
			expect: "[2019-11-03 11:12:13.000000] ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(formatTimestamp(tc.input))
			if got != tc.expect {
				t.Errorf("something went wrong. expect: %s, got: %s", tc.expect, got)
			}
		})
	}
}
