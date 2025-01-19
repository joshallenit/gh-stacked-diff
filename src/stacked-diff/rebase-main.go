package stacked_diff

import (
	"log"
	ex "stacked-diff-workflow/src/execute"
	"strings"
)

func RebaseMain(logger *log.Logger) {
	RequireMainBranch()
	Stash("rebase-main")

	ex.ExecuteOrDie(ex.ExecuteOptions{
		Output: ex.NewStandardOutput(),
	}, "git", "fetch")
	username := GetUsername()
	originSummaries := strings.Split(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+ex.GetMainBranch(), "-n", "30", "--format=%s", "--author="+username)), "\n")
	localLogs := strings.Split(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+ex.GetMainBranch()+"..HEAD", "--format=%h %s")), "\n")
	// Look for matching summaries between localCommits and originCommits
	var dropCommits []string

	for _, localLog := range localLogs {
		localCommit := localLog[0:strings.Index(localLog, " ")]
		println("LOCAL commit", localCommit)
		localSummary := localLog[len(localCommit)+1:]
		println("LOCAL summary", localSummary)
		if contains(originSummaries, localSummary) {
			logger.Println("Force dropping as it was already merged:", localCommit, localSummary)
			dropCommits = append(dropCommits, localCommit)
		}
	}

	if len(dropCommits) > 0 {
		environmentVariables := []string{"GIT_SEQUENCE_EDITOR=sequence-editor-drop-already-merged " + strings.Join(dropCommits, " ")}
		options := ex.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			Output:               ex.NewStandardOutput(),
		}
		ex.ExecuteOrDie(options, "git", "rebase", "-i", "origin/"+ex.GetMainBranch())
	} else {
		options := ex.ExecuteOptions{
			Output: ex.NewStandardOutput(),
		}
		ex.ExecuteOrDie(options, "git", "rebase", "origin/"+ex.GetMainBranch())
	}
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
