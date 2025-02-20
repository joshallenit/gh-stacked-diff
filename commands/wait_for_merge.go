package commands

import (
	"log/slog"
	ex "stackeddiff/execute"
	"strings"
	"time"
)

// Waits for a pull request to be merged.
func WaitForMerge(commitIndicator string, indicatorType IndicatorType, silent bool) {
	branchName := GetBranchInfo(commitIndicator, indicatorType).BranchName
	for getMergedAt(branchName) == "" {
		slog.Info("Not merged yet...")
		ex.Sleep(30 * time.Second)
	}
	slog.Info("Merged!")
	if !silent {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "say", "P R has been merged")
	}
}

func getMergedAt(branchName string) string {
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "mergedAt", "--jq", ".mergedAt"))
}
