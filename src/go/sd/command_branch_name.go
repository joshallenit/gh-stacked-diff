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
		fmt.Fprintln(os.Stderr, "Outputs the branch name for a given commit hash or pull request number. Useful for custom scripting.")
		fmt.Fprintln(os.Stderr, "sd branch-name [flags] <commitIndicator>")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "Missing commitIndicator")
			flagSet.Usage()
			os.Exit(1)
		}
		indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
		branchName := sd.GetBranchInfo(flagSet.Arg(0), indicatorType).BranchName
		fmt.Fprint(stdOut, branchName)
	}}
}
