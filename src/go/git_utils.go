package stackeddiff

import (
	_ "embed"
	"log/slog"
	"slices"
	"strings"

	ex "stackeddiff/execute"
)

type GitLog struct {
	Commit  string
	Subject string
	Branch  string
}

var mainBranchName string

func GetMainBranch() string {
	if mainBranchName == "" {
		remoteMainBranch, err := ex.Execute(ex.ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "origin/HEAD")
		if err == nil {
			remoteMainBranch = strings.TrimSpace(remoteMainBranch)
			mainBranchName = remoteMainBranch[strings.Index(remoteMainBranch, "/")+1:]
		} else {
			// Remote is empty, or the repository was not cloned, use config.
			mainBranchName = strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "init.defaultBranch"))
			if !LocalHasBranch(mainBranchName) {
				panic("Cannot determine name of main branch.\n" +
					"Use \"git remote set-head origin main\" to set the name to main.")
			}
		}
	}
	return mainBranchName
}

func newGitLogs(logsRaw string) []GitLog {
	logLines := strings.Split(strings.TrimSpace(logsRaw), "\n")
	var logs []GitLog
	for _, logLine := range logLines {
		components := strings.Split(logLine, formatDelimiter)
		if len(components) != 3 {
			// No git logs.
			continue
		}
		logs = append(logs, GitLog{Commit: components[0], Subject: components[1], Branch: GetBranchForSantizedSubject(components[2])})
	}
	return logs
}

func GetAllCommits() []GitLog {
	gitArgs := []string{"--no-pager", "log", "--pretty=format:%h" + formatDelimiter + "%s" + formatDelimiter + "%f", "--abbrev-commit"}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

func GetNewCommits(compareFromRemoteBranch string, to string) []GitLog {
	gitArgs := []string{"--no-pager", "log", "--pretty=format:%h" + formatDelimiter + "%s" + formatDelimiter + "%f", "--abbrev-commit"}
	if RemoteHasBranch(compareFromRemoteBranch) {
		gitArgs = append(gitArgs, "origin/"+compareFromRemoteBranch+".."+to)
	} else {
		gitArgs = append(gitArgs, to)
	}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

// Returns most recent commit of the given branch that is on origin/main, or "" if the main branch is not on remote.
func FirstOriginMainCommit(branchName string) string {
	if !LocalHasBranch(branchName) {
		panic("Branch does not exist " + branchName)
	}
	// Verify that remote has branch, there is no origin commit.
	if !RemoteHasBranch(GetMainBranch()) {
		return ""
	}
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "merge-base", "origin/"+GetMainBranch(), branchName))
}

func RemoteHasBranch(branchName string) bool {
	remoteBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-r", "--list", "origin/"+branchName))
	return remoteBranch != ""
}

func LocalHasBranch(branchName string) bool {
	localBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "--list", branchName))
	return localBranch != ""
}

func RequireMainBranch() {
	if GetCurrentBranchName() != GetMainBranch() {
		panic("Must be run from " + GetMainBranch() + " branch")
	}
}

func RequireCommitOnMain(commit string) {
	if commit == GetMainBranch() {
		return
	}
	newCommits := GetNewCommits(GetMainBranch(), "HEAD")
	if !slices.ContainsFunc(newCommits, func(gitLog GitLog) bool {
		return gitLog.Commit == commit
	}) {
		panic("Commit " + commit + " does not exist on " + GetMainBranch() + ". Check `sd log` for available commits.")
	}
}

func GetCurrentBranchName() string {
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "HEAD"))
}

func Stash(forName string) bool {
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before "+forName))
	if strings.HasPrefix(stashResult, "Saved working") {
		slog.Info(stashResult)
		return true
	}
	return false
}

func PopStash(popStash bool) {
	if popStash {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "pop")
		slog.Info("Popped stash back")
	}
}
