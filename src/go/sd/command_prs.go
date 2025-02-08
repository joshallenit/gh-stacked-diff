package main

import (
	"flag"
	"io"
	"log/slog"
	ex "stackeddiff/execute"
)

func createPrsCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("prs", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Lists all Pull Requests you have open.",
		Description: "Lists all Pull Requests you have open.\n" +
			"\n" +
			"You must be logged-in, via \"gh auth login\"",
		Usage:           "sd " + flagSet.Name(),
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(command Command) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			ex.ExecuteOrDie(ex.ExecuteOptions{Output: &ex.ExecutionOutput{Stdout: stdOut, Stderr: stdOut}},
				"gh", "pr", "list", "--author", "@me")
		}}
}
