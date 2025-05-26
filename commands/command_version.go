package commands

import (
	"flag"
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
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(asyncConfig.App, flagSet, "too many args", command.Usage)
			}
			var stableSuffix string
			if util.CurrentVersion == util.StableVersion {
				stableSuffix = " (stable)"
			} else {
				stableSuffix = " (preview)"
			}
			util.Fprintln(asyncConfig.App.Io.Out, "Version "+util.CurrentVersion+stableSuffix)
		}}
}
