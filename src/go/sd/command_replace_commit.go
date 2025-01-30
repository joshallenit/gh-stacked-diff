package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	sd "stackeddiff"
)

func CreateReplaceCommitCommand() Command {
	flagSet := flag.NewFlagSet("replace-commit", flag.ExitOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Replaces a commit on current branch with the contents of another branch")
		fmt.Fprintln(os.Stderr, "sd replace-commit [flags] <commitIndicator>")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "Missing commitIndicator")
			flagSet.Usage()
			os.Exit(1)
		}
		indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
		sd.ReplaceCommit(flagSet.Arg(0), indicatorType)
	}}
}
