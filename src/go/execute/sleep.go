package execute

import (
	"time"
)

var defaultSleep = time.Sleep

/*
Sleep function that can be overridden for testing.
Default is [time.Sleep]
*/
func Sleep(d time.Duration) {
	defaultSleep(d)
}

func SetDefaultSleep(newSleep func(d time.Duration)) {
	defaultSleep = newSleep
}
