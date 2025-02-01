package main

import (
	"flag"
	"fmt"
	"os"
	sd "stackeddiff"
)

func CreateWaitForMergeCommand() Command {
	flagSet := flag.NewFlagSet("wait-for-merge", flag.ExitOnError)

	indicatorTypeString := AddIndicatorFlag(flagSet)

	var silent bool
	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output")

	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Waits for a pull request to be merged. Polls PR every 30 seconds. Useful for your own custom scripting.")
		fmt.Fprintln(flagSet.Output(), "sd wait-for-merge [flags] <commit hash or pull request number>")
		flagSet.PrintDefaults()
	}

	return Command{
		FlagSet:      flagSet,
		UsageSummary: "Waits for a pull request to be merged",
		OnSelected: func() {
			if flagSet.NArg() != 1 {
				flagSet.Usage()
				os.Exit(1)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			sd.WaitForMerge(flagSet.Arg(0), indicatorType, silent)
		}}
}
