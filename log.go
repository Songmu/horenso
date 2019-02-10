package horenso

import (
	"fmt"
	"log"
	"strings"
)

type loglevel int

const (
	mute loglevel = iota
	warn
	info
	debug
)

func (ho *horenso) logLevel() loglevel {
	return loglevel(len(ho.Verbose))
}

func (ho *horenso) logf(lv loglevel, format string, a ...interface{}) {
	ho.log(lv, fmt.Sprintf(format, a...))
}

func (ho *horenso) log(lv loglevel, str string) {
	if ho.logLevel() < lv || lv <= mute {
		return
	}
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	log.Print(str)
}
