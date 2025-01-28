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
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Replaces a commit on current branch with the contents of another branch")
		fmt.Fprintln(os.Stderr, "sd replace-commit <commit hash or PR number>")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		sd.ReplaceCommit(flagSet.Arg(0))
	}}
}
