package main

import (
	"flag"
	"io"
	"log/slog"
	sd "stackeddiff"
)

func CreateLogCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("log", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Displays git log of your changes",
		Description: "Displays summary of the git commits on current branch that are not in remote branch.\n" +
			"Useful to view list indexes, or copy commit hashes, to use for the other commands.",
		Usage:           "sd " + flagSet.Name(),
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(command Command) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			sd.PrintGitLog(stdOut)
		}}
}
