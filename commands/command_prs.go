package commands

import (
	"flag"
	"log/slog"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
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
		OnSelected: func(appConfig util.AppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(appConfig, flagSet, "too many arguments", command.Usage)
			}
			ex.ExecuteOrDie(ex.ExecuteOptions{Output: &ex.ExecutionOutput{Stdout: appConfig.Io.Out, Stderr: appConfig.Io.Err}},
				"gh", "pr", "list", "--author", "@me")
		}}
}
