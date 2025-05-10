package commands

import (
	"flag"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createPrsCommand() Command {
	flagSet := flag.NewFlagSet("prs", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Lists all Pull Requests you have open.",
		Description: "Lists all Pull Requests you have open.\n" +
			"\n" +
			"You must be logged-in, via \"gh auth login\"",
		Usage:           "sd " + flagSet.Name(),
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(asyncConfig.App, flagSet, "too many arguments", command.Usage)
			}
			util.ExecuteOrDie(util.ExecuteOptions{Io: asyncConfig.App.Io},
				"gh", "pr", "list", "--author", "@me")
		}}
}
