package main

import (
	"flag"
	"log/slog"
)

type Command struct {
	FlagSet    *flag.FlagSet
	OnSelected func()
	// Default if not set is 0 which is Info.
	DefaultLogLevel slog.Level
}
