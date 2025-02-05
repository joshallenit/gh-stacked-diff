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
		FlagSet: flagSet,
		Summary: "Checks out branch associated with commit indicator",
		Description: "Checks out the branch associated with commit indicator.\n" +
			"\n" +
			"For when you want to merge only the branch with with origin/" + sd.GetMainBranch() + ",\n" +
			"rather than your entire local " + sd.GetMainBranch() + " branch, verify why \n" +
			"CI is failing on that particular branch, or for any other reason.\n" +
			"\n" +
			"After modifying the branch you can use \"sd replace-commit\" to sync local " + sd.GetMainBranch() + ".",
		Usage: "sd " + flagSet.Name() + " [flags] <commitIndicator>",
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
