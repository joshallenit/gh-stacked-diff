package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	sd "stackeddiff"
)

func CreateBranchNameCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("branch-name", flag.ExitOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Outputs the branch name for a given commit indicator. Useful for your own custom scripting.")
		fmt.Fprintln(flagSet.Output(), "sd branch-name [flags] <commitIndicator>")
		flagSet.PrintDefaults()
	}

	return Command{
		FlagSet:         flagSet,
		DefaultLogLevel: slog.LevelError,
		UsageSummary:    "Outputs branch name of commit indicator",
		OnSelected: func() {
			if flagSet.NArg() != 1 {
				fmt.Fprintln(flagSet.Output(), "Missing commitIndicator")
				flagSet.Usage()
				os.Exit(1)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			branchName := sd.GetBranchInfo(flagSet.Arg(0), indicatorType).BranchName
			fmt.Fprint(stdOut, branchName)
		}}
}
