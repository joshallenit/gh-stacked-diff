package main

import (
	"flag"
	"io"
	"log/slog"
	sd "stackeddiff"
)

func createLogCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("log", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Displays git log of your changes",
		Description: "Displays summary of the git commits on current branch that are not\n" +
			"in the remote branch.\n" +
			"\n" +
			"Useful to view list indexes, or copy commit hashes, to use for the\n" +
			"commitIndicator required by other commands.\n" +
			"\n" +
			"A ✅ means that there is a PR associated with the commit (actually it\n" +
			"means there is a branch, but having a branch means there is a PR when\n" +
			"using this workflow). If there is more than one commit on the\n" +
			"associated branch, those commits are also listed (indented under the\n" +
			"their associated commit summary).",
		Usage:           "sd " + flagSet.Name(),
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(command Command) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			sd.PrintGitLog(stdOut)
		}}
}
