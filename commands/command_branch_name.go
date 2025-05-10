package commands

import (
	"flag"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
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
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() > 1 {
				commandError(asyncConfig.App, flagSet, "too many arguments", command.Usage)
			}
			selectCommitOptions := interactive.CommitSelectionOptions{
				Prompt:      "What commit do you want the branch name for?",
				CommitType:  interactive.CommitTypeBoth,
				MultiSelect: false,
			}
			targetCommit := getTargetCommits(asyncConfig.App, command, []string{flagSet.Arg(0)}, indicatorTypeString, selectCommitOptions)
			util.Fprint(asyncConfig.App.Io.Out, targetCommit[0].Branch)
		}}
}
