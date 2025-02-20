package commands

import (
	"flag"
	sd "stackeddiff"
)

func createWaitForMergeCommand() Command {
	flagSet := flag.NewFlagSet("wait-for-merge", flag.ContinueOnError)

	indicatorTypeString := addIndicatorFlag(flagSet)
	silent := addSilentFlag(flagSet, "")

	return Command{
		FlagSet: flagSet,
		Summary: "Waits for a pull request to be merged",
		Description: "Waits for a pull request to be merged. Polls PR every 30 seconds.\n" +
			"\n" +
			"Useful for your own custom scripting.",
		Usage: "sd " + flagSet.Name() + " [flags] <commit hash or pull request number>",
		OnSelected: func(command Command) {
			if flagSet.NArg() != 1 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			sd.WaitForMerge(flagSet.Arg(0), indicatorType, *silent)
		}}
}
