package main

import (
	"flag"
	"fmt"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateCheckoutCommand() Command {
	flagSet := flag.NewFlagSet("checkout", flag.ExitOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Checks out the branch associated with the PR or commit.")
		fmt.Fprintln(flagSet.Output(), "sd checkout [flags] <commitIndicator>")
		flagSet.PrintDefaults()
	}

	return Command{
		FlagSet:      flagSet,
		UsageSummary: "Checks out branch associated with commit indicator",
		OnSelected: func() {
			if flagSet.NArg() != 1 {
				flagSet.Usage()
				os.Exit(1)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			branchName := sd.GetBranchInfo(flagSet.Arg(0), indicatorType).BranchName
			ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "checkout", branchName)
		}}
}
