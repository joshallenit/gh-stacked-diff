package commands

import (
	"flag"
	"log/slog"
)

// sd program command.
type Command struct {
	// Any flags specific for the command.
	FlagSet *flag.FlagSet
	// Executes the command.
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
