package commands

import (
	"fmt"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
)

// Guaranteed to return at least one value (or else appConfig.Exit will be called).
func getTargetCommits(
	appConfig util.AppConfig,
	command Command,
	// Note: empty values are ignored for convienience to allow use of FlagSet.Arg().
	commitsFromCommandLine []string,
	indicatorTypeString *string,
	options interactive.CommitSelectionOptions,
) []templates.GitLog {
	commitsFromCommandLine = util.FilterSlice(commitsFromCommandLine, func(commit string) bool {
		return commit != ""
	})
	if len(commitsFromCommandLine) == 0 {
		if !interactive.IsOutputTerminal(appConfig.Io.Out) || !interactive.IsOutputTerminal(appConfig.Io.Err) {
			commandError(command.FlagSet, "Target commit not specified and cannot ask interactively because not a terminal", command.Usage)
		}
		selectedCommits, err := interactive.GetCommitSelection(appConfig.Io, options)
		if err != nil {
			commandError(command.FlagSet, err.Error(), command.Usage)
		}
		if len(selectedCommits) == 0 {
			appConfig.Exit(nil)
		}
		slog.Info("Target commits: " + fmt.Sprint(selectedCommits))
		return selectedCommits
	} else {
		indicatorType := checkIndicatorFlag(command, indicatorTypeString)
		return util.MapSlice(commitsFromCommandLine, func(commit string) templates.GitLog {
			return templates.GetBranchInfo(commit, indicatorType)
		})
	}
}
