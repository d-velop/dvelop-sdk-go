package structuredlog

import (
	"time"
)

type Clock func() time.Time

var clock Clock = time.Now

func SetClock(c Clock) {
	clock = c
}

