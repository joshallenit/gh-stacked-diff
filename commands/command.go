/*
Commands of the the Stacked Diff Workflow Command Line Interface.
*/
package commands

import (
	"flag"
	"io"
	"log/slog"
)

// sd program command.
type Command struct {
	// Any flags specific for the command.
	FlagSet *flag.FlagSet
	// Executes the command.
	OnSelected func(command Command, stdOut io.Writer, stdErr io.Writer, sequenceEditorPrefix string, exit func(err any))
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
