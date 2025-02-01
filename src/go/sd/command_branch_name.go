package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	sd "stackeddiff"
)

func CreateBranchNameCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("branch-name", flag.ContinueOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	return Command{
		FlagSet:         flagSet,
		DefaultLogLevel: slog.LevelError,
		Summary:         "Outputs branch name of commit",
		Description:     "Outputs the branch name for a given commit indicator. Useful for your own custom scripting.",
		Usage:           "sd " + flagSet.Name() + " [flags] <commitIndicator>",
		OnSelected: func(command Command) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			branchName := sd.GetBranchInfo(flagSet.Arg(0), indicatorType).BranchName
			fmt.Fprint(stdOut, branchName)
		}}
}
