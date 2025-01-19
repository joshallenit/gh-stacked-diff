package main

import (
	"flag"
	"log"
	sd "stacked-diff-workflow/src/stacked-diff"
)

/*
Find out if any of the commits have already been merged and automatically drop
them to avoid having to deal with merge conflicts that have already been fixed
in main.
*/
func main() {
	var logFlags int
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Parse()
	log.SetFlags(logFlags)
	sd.RebaseMain(log.Default())
}
