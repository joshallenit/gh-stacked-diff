package commands

import (
	"flag"
	"io"
	"log/slog"
	"strings"
	"time"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
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
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, sequenceEditorPrefix string, exit func(err any)) {
			if flagSet.NArg() != 1 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			waitForMerge(flagSet.Arg(0), indicatorType, *silent)
		}}
}

// Waits for a pull request to be merged.
func waitForMerge(commitIndicator string, indicatorType templates.IndicatorType, silent bool) {
	branchName := templates.GetBranchInfo(commitIndicator, indicatorType).Branch
	for getMergedAt(branchName) == "" {
		slog.Info("Not merged yet...")
		util.Sleep(30 * time.Second)
	}
	slog.Info("Merged!")
	if !silent {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "say", "P R has been merged")
	}
}

func getMergedAt(branchName string) string {
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "mergedAt", "--jq", ".mergedAt"))
}
