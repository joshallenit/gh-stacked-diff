package commands

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createBranchNameCommand() Command {
	flagSet := flag.NewFlagSet("branch-name", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)
	return Command{
		FlagSet:         flagSet,
		DefaultLogLevel: slog.LevelError,
		Summary:         "Outputs branch name of commit",
		Description: "Outputs the branch name for a given commit indicator.\n" +
			"Useful for your own custom scripting.",
		Usage: "sd " + flagSet.Name() + " [flags] <commitIndicator>",
		OnSelected: func(appConfig util.AppConfig, command Command) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			branchName := templates.GetBranchInfo(flagSet.Arg(0), indicatorType).Branch
			fmt.Fprint(appConfig.Io.Out, branchName)
		}}
}
