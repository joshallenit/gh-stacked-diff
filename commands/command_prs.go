package commands

import (
	"flag"
	"io"
	"log/slog"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
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
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, sequenceEditorPrefix string, exit func(err any)) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			ex.ExecuteOrDie(ex.ExecuteOptions{Output: &ex.ExecutionOutput{Stdout: stdOut, Stderr: stdErr}},
				"gh", "pr", "list", "--author", "@me")
		}}
}
