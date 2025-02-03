package main

import (
	"flag"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateCheckoutCommand() Command {
	flagSet := flag.NewFlagSet("checkout", flag.ContinueOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	return Command{
		FlagSet:     flagSet,
		Summary:     "Checks out branch associated with commit indicator",
		Description: "Checks out the branch associated with the PR or commit.",
		Usage:       "sd " + flagSet.Name() + " [flags] <commitIndicator>",
		OnSelected: func(command Command) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			indicatorType := CheckIndicatorFlag(command, indicatorTypeString)
			branchName := sd.GetBranchInfo(flagSet.Arg(0), indicatorType).BranchName
			ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "checkout", branchName)
		}}
}
