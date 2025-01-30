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
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Outputs the branch name for a given commit hash or pull request number. Useful for custom scripting.")
		fmt.Fprintln(os.Stderr, "sd branch-name <commit hash or pull request number>")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		branchName := sd.GetBranchInfo(flagSet.Arg(0), sd.IndicatorTypeGuess).BranchName
		fmt.Fprint(stdOut, branchName)
	}}
}
