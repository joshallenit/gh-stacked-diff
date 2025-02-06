package stackeddiff

import (
	"fmt"
	"log/slog"
	"strings"

	ex "stackeddiff/execute"
)

func RebaseMain() {
	requireMainBranch()
	shouldPopStash := stash("rebase-main")

	slog.Info("Fetching...")
	ex.ExecuteOrDie(ex.ExecuteOptions{
		Output: ex.NewStandardOutput(),
	}, "git", "fetch")
	username := GetUsername()
	originSummaries := strings.Split(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+GetMainBranchOrDie(), "-n", "30", "--format=%s", "--author="+username)), "\n")
	localLogs := strings.Split(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+GetMainBranchOrDie()+"..HEAD", "--format=%h %s")), "\n")
	// Look for matching summaries between localCommits and originCommits
	var dropCommits []string

	for _, localLog := range localLogs {
		spaceIndex := strings.Index(localLog, " ")
		if spaceIndex == -1 {
			slog.Info("No local changes")
			break
		}
		localCommit := localLog[0:spaceIndex]
		localSummary := localLog[len(localCommit)+1:]
		if contains(originSummaries, localSummary) {
			slog.Info(fmt.Sprint("Force dropping as it was already merged: ", localCommit, " ", localSummary))
			dropCommits = append(dropCommits, localCommit)
		}
	}

	slog.Info("Rebasing...")
	if len(dropCommits) > 0 {
		environmentVariables := []string{
			"GIT_SEQUENCE_EDITOR=sequence_editor_drop_already_merged " +
				strings.Join(dropCommits, " "),
		}
		options := ex.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			Output:               ex.NewStandardOutput(),
		}
		ex.ExecuteOrDie(options, "git", "rebase", "-i", "origin/"+GetMainBranchOrDie())
	} else {
		options := ex.ExecuteOptions{
			Output: ex.NewStandardOutput(),
		}
		ex.ExecuteOrDie(options, "git", "rebase", "origin/"+GetMainBranchOrDie())
	}
	popStash(shouldPopStash)
}

func contains(originSummaries []string, localSummary string) bool {
	for _, originSummary := range originSummaries {
		// Github will add a " (#pr_number)" to merged PR commit summaries.
		if localSummary == originSummary || strings.HasPrefix(originSummary, localSummary+" (#") {
			return true
		}
	}
	return false
}
