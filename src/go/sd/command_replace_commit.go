package main

import (
	"flag"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateReplaceCommitCommand() Command {
	flagSet := flag.NewFlagSet("replace-commit", flag.ContinueOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	return Command{
		FlagSet: flagSet,
		Summary: "Replaces a commit on " + ex.GetMainBranch() + " branch with the contents its associated branch",
		Description: "Replaces a commit on " + ex.GetMainBranch() + " branch with the contents its associated branch.\n" +
			"This is useful when you make changes within a branch, for example to fix a problem found on CI,\n" +
			"and want to bring the changes over to your local " + ex.GetMainBranch() + " branch.",
		Usage: "sd replace-commit [flags] <commitIndicator>",
		OnSelected: func(command Command) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			sd.ReplaceCommit(flagSet.Arg(0), indicatorType)
		}}
}
