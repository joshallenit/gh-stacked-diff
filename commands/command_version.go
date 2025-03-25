package commands

import (
	"flag"
	"fmt"
	"io"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createVersionCommand() Command {
	flagSet := flag.NewFlagSet("version", flag.ContinueOnError)
	return Command{
		FlagSet:         flagSet,
		DefaultLogLevel: slog.LevelError,
		Summary:         "Outputs version number",
		Description:     "Outputs the version number.",
		Usage:           "sd " + flagSet.Name(),
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, sequenceEditorPrefix string, exit func(err any)) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many args", command.Usage)
			}
			var stableSuffix string
			if util.CurrentVersion == util.StableVersion {
				stableSuffix = " (stable)"
			} else {
				stableSuffix = " (preview)"
			}
			fmt.Fprint(stdOut, "Version "+util.CurrentVersion+stableSuffix)
		}}
}
