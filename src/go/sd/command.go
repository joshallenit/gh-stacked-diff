package main

import (
	"flag"
	"log/slog"
)

type Command struct {
	FlagSet    *flag.FlagSet
	OnSelected func(command Command)
	// Default if not set is 0 which is Info.
	DefaultLogLevel slog.Level
	// Short, one line summary of command
	Summary string
	// Longer descripton of command
	Description string
	// Usage to print out during help and if wrong arguments to command
	Usage string
}
