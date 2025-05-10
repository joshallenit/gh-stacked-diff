package commands

import (
	"flag"
	"log/slog"
	"strings"
	"time"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
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
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() > 1 {
				commandError(asyncConfig.App, flagSet, "too many arguments", command.Usage)
			}
			selectCommitOptions := interactive.CommitSelectionOptions{
				Prompt:      "What PR do you want to wait for to be merged?",
				CommitType:  interactive.CommitTypePr,
				MultiSelect: false,
			}
			targetCommit := getTargetCommits(asyncConfig.App, command, []string{flagSet.Arg(0)}, indicatorTypeString, selectCommitOptions)
			waitForMerge(targetCommit[0], *silent)
		}}
}

// Waits for a pull request to be merged.
func waitForMerge(targetCommit templates.GitLog, silent bool) {
	for getMergedAt(targetCommit.Branch) == "" {
		slog.Info("Not merged yet...")
		util.Sleep(30 * time.Second)
	}
	slog.Info("Merged!")
	if !silent {
		util.ExecuteOrDie(util.ExecuteOptions{}, "say", "P R has been merged")
	}
}

func getMergedAt(branchName string) string {
	return strings.TrimSpace(util.ExecuteOrDie(util.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "mergedAt", "--jq", ".mergedAt"))
}
