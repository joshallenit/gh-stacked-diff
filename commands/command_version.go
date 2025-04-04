package commands

import (
	"flag"
	"fmt"
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
		OnSelected: func(appConfig util.AppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(appConfig, flagSet, "too many args", command.Usage)
			}
			var stableSuffix string
			if util.CurrentVersion == util.StableVersion {
				stableSuffix = " (stable)"
			} else {
				stableSuffix = " (preview)"
			}
			if _, err := fmt.Fprint(appConfig.Io.Out, "Version "+util.CurrentVersion+stableSuffix); err != nil {
				panic(err)
			}
		}}
}
