/*
Commands of the the Stacked Diff Workflow Command Line Interface.
*/
package commands

import (
	"flag"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

type OnSelectedFunc func(asyncConfig util.AsyncAppConfig, command Command)

// sd program command.
type Command struct {
	// Any flags specific for the command.
	FlagSet *flag.FlagSet
	// Executes the command.
	OnSelected OnSelectedFunc
	// Default if not set is 0 which is Info.
	DefaultLogLevel slog.Level
	// Short, one line summary of command
	Summary string
	// Longer descripton of command
	Description string
	// Usage to print out during help and if wrong arguments to command
	Usage string
	// Whether to include in help messages.
	Hidden bool
}
