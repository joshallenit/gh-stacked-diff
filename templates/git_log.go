package templates

import (
	"slices"
	"strings"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

// Returned by some of the Get*Commit functions.
type GitLog struct {
	// Abbreviated commit hash.
	Commit string
	// Full commit hash.
	CommitFull string
	// Commit subject.
	Subject string
	// Associated branch name. Branch might not exist.
	Branch string
}

// Delimter for git log format when a space cannot be used.
const formatDelimiter = "|stackeddiff-delim|"

// Format sent to "git log" for use by [newGitLogs].
const newGitLogsFormat = "--pretty=format:%h" + formatDelimiter + "%H" + formatDelimiter + "%s" + formatDelimiter + "%f"

// Returns all the commits on the current branch. For use by tests.
func GetAllCommits() []GitLog {
	gitArgs := []string{"--no-pager", "log", newGitLogsFormat, "--abbrev-commit"}
	logsRaw := util.ExecuteOrDie(util.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

func GetNewCommits(to string) []GitLog {
	compareFromRemoteBranch := util.GetMainBranchOrDie()
	gitArgs := []string{"--no-pager", "log", newGitLogsFormat, "--abbrev-commit"}
	if util.RemoteHasBranch(compareFromRemoteBranch) {
		gitArgs = append(gitArgs, "origin/"+compareFromRemoteBranch+".."+to)
	} else {
		gitArgs = append(gitArgs, to)
	}
	logsRaw := util.ExecuteOrDie(util.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

func newGitLogs(logsRaw string) []GitLog {
	logLines := strings.Split(strings.TrimSpace(logsRaw), "\n")
	var logs []GitLog
	for _, logLine := range logLines {
		components := strings.Split(logLine, formatDelimiter)
		if len(components) != 4 {
			// No git logs.
			continue
		}
		logs = append(logs, GitLog{Commit: components[0], CommitFull: components[1], Subject: components[2], Branch: getBranchForSantizedSubject(components[3])})
	}
	return logs
}

func RequireCommitOnMain(commit string) {
	if commit == util.GetMainBranchOrDie() {
		return
	}
	newCommits := GetNewCommits("HEAD")
	if !slices.ContainsFunc(newCommits, func(gitLog GitLog) bool {
		return gitLog.Commit == commit
	}) {
		panic("Commit " + commit + " does not exist on " + util.GetMainBranchOrDie() + ". Check `sd log` for available commits.")
	}
}
