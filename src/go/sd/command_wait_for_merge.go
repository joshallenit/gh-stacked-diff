package main

import (
	"flag"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateWaitForMergeCommand() Command {
	flagSet := flag.NewFlagSet("wait-for-merge", flag.ContinueOnError)

	indicatorTypeString := AddIndicatorFlag(flagSet)
	silent := AddSilentFlag(flagSet, "")

	return Command{
		FlagSet:     flagSet,
		Summary:     "Waits for a pull request to be merged",
		Description: "Waits for a pull request to be merged. Polls PR every 30 seconds. Useful for your own custom scripting.",
		Usage:       "sd " + ex.GetMainBranch() + " [flags] <commit hash or pull request number>",
		OnSelected: func(command Command) {
			if flagSet.NArg() != 1 {
				flagSet.Usage()
				os.Exit(1)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			sd.WaitForMerge(flagSet.Arg(0), indicatorType, *silent)
		}}
}
