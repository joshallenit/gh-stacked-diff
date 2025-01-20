package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	sd "stacked-diff-workflow/src/stacked-diff"
)

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {
	var logFlags int
	var logLevelString string
	flag.IntVar(&logFlags, "log-flags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.StringVar(&logLevelString, "log-level", "info", "Log level: debug, info, warn, or error")
	flag.Parse()
	log.SetFlags(logFlags)
	var logLevel slog.Level
	var unmarshallErr = logLevel.UnmarshalText([]byte(logLevelString))
	if unmarshallErr != nil {
		panic("Invalid log level " + logLevelString + ": " + unmarshallErr.Error())
	}
	slog.SetLogLoggerLevel(logLevel)
	sd.PrintGitLog(os.Stdout)
}
