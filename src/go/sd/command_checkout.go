package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateCheckoutCommand() Command {
	flagSet := flag.NewFlagSet("checkout", flag.ExitOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Checks out the branch associated with the PR or commit.")
		fmt.Fprintln(os.Stderr, "sd checkout [flags] <commitIndicator>")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
		branchName := sd.GetBranchInfo(flagSet.Arg(0), indicatorType).BranchName
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "checkout", branchName)
	}}
}
