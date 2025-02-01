package main

import (
	"flag"
	"fmt"
	"os"
	sd "stackeddiff"
)

func CreateReplaceCommitCommand() Command {
	flagSet := flag.NewFlagSet("replace-commit", flag.ExitOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Replaces a commit on current branch with the contents of another branch")
		fmt.Fprintln(flagSet.Output(), "sd replace-commit [flags] <commitIndicator>")
		flagSet.PrintDefaults()
	}

	return Command{
		FlagSet:      flagSet,
		UsageSummary: "Replaces a commit on main branch with the contents its associated branch",
		OnSelected: func() {
			if flagSet.NArg() != 1 {
				fmt.Fprintln(flagSet.Output(), "error: missing commitIndicator")
				flagSet.Usage()
				os.Exit(1)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			sd.ReplaceCommit(flagSet.Arg(0), indicatorType)
		}}
}
