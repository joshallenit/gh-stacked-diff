package main

import (
	"flag"
	sd "stackeddiff"
)

func CreateWaitForMergeCommand() Command {
	flagSet := flag.NewFlagSet("wait-for-merge", flag.ContinueOnError)

	indicatorTypeString := AddIndicatorFlag(flagSet)
	silent := AddSilentFlag(flagSet, "")

	return Command{
		FlagSet: flagSet,
		Summary: "Waits for a pull request to be merged",
		Description: "Waits for a pull request to be merged. Polls PR every 30 seconds.\n" +
			"\n" +
			"Useful for your own custom scripting.",
		Usage: "sd " + sd.GetMainBranch() + " [flags] <commit hash or pull request number>",
		OnSelected: func(command Command) {
			if flagSet.NArg() != 1 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			indicatorType := CheckIndicatorFlag(command, indicatorTypeString)
			sd.WaitForMerge(flagSet.Arg(0), indicatorType, *silent)
		}}
}
