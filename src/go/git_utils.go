package stackeddiff

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	ex "stackeddiff/execute"
)

// Returned by some of the Get*Commit functions.
type GitLog struct {
	// Abbreviated commit hash.
	Commit string
	// Commit subject.
	Subject string
	// Associated branch name. Branch might not exist.
	Branch string
}

// Cached value of main branch name.
var mainBranchName string

// Cached value of user email.
var userEmail string

// Delimter for git log format when a space cannot be used.
const formatDelimiter = "|stackeddiff-delim|"

// Format sent to "git log" for use by [newGitLogs].
const newGitLogsFormat = "--pretty=format:%h" + formatDelimiter + "%s" + formatDelimiter + "%f"

// Returns name of main branch, or panics if cannot be determined.
func GetMainBranchOrDie() string {
	mainBranch, err := getMainBranch()
	if err != nil {
		panic(fmt.Sprint("Could not get main branch: ", err))
	}
	return mainBranch
}

// Returns name of main branch, or "main" if cannot be determined. For use by CLI help.
func GetMainBranchForHelp() string {
	mainBranch, err := getMainBranch()
	if err != nil {
		return "main"
	}
	return mainBranch
}

func getMainBranch() (string, error) {
	if mainBranchName == "" {
		remoteMainBranch, err := ex.Execute(ex.ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "origin/HEAD")
		if err == nil {
			remoteMainBranch = strings.TrimSpace(remoteMainBranch)
			mainBranchName = remoteMainBranch[strings.Index(remoteMainBranch, "/")+1:]
		} else {
			// Check possible reasons.
			_, inRepositoryErr := ex.Execute(ex.ExecuteOptions{}, "git", "rev-parse")
			if inRepositoryErr != nil {
				return "", errors.New("not in a git repository. Must be run from a git repository")
			}
			_, hasHeadErr := ex.Execute(ex.ExecuteOptions{}, "git", "rev-list", "--max-parents=0", "HEAD")
			if hasHeadErr != nil {
				return "", errors.New("the remote repository is empty.\n" +
					"Push an initial inconsequential commit to origin/main and try again. \n" +
					"Using a repository without an initial remote commit is not recommended because git \n" +
					"requires special handling for the root commit, and you might accidentially \n" +
					"create more than one root commit")
			}
			// Try to set the remote head and try again.
			trySetRemoteHead()
			remoteMainBranch, err := ex.Execute(ex.ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "origin/HEAD")
			if err == nil {
				remoteMainBranch = strings.TrimSpace(remoteMainBranch)
				mainBranchName = remoteMainBranch[strings.Index(remoteMainBranch, "/")+1:]
			} else {
				// else the repository was not cloned
				return "", errors.New("cannot determine name of main branch because remote head is not set.\n" +
					"This usually means that repository was not cloned (git init was used instead). \n" +
					"Set the name of the main branch on remote and try again:\n" +
					"git remote set-head origin main")
			}
		}
	}
	return mainBranchName, nil
}

func trySetRemoteHead() {
	defer func() { _ = recover() }()
	currentBranch := GetCurrentBranchName()
	defaultBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "init.defaultBranch"))
	if currentBranch == defaultBranch || currentBranch == "main" {
		slog.Warn("Setting remote head to " + currentBranch + " because it was not set.")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "remote", "set-head", "origin", currentBranch)
	}
}

func getUsername() string {
	if userEmail == "" {
		userEmailRaw := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "user.email"))
		userEmail = userEmailRaw[0:strings.Index(userEmailRaw, "@")]
	}
	return userEmail
}

// Returns all the commits on the current branch. For use by tests.
func GetAllCommits() []GitLog {
	gitArgs := []string{"--no-pager", "log", newGitLogsFormat, "--abbrev-commit"}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

func getNewCommits(to string) []GitLog {
	compareFromRemoteBranch := GetMainBranchOrDie()
	gitArgs := []string{"--no-pager", "log", newGitLogsFormat, "--abbrev-commit"}
	if RemoteHasBranch(compareFromRemoteBranch) {
		gitArgs = append(gitArgs, "origin/"+compareFromRemoteBranch+".."+to)
	} else {
		gitArgs = append(gitArgs, to)
	}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
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
		logs = append(logs, GitLog{Commit: components[0], Subject: components[1], Branch: getBranchForSantizedSubject(components[2])})
	}
	return logs
}

// Returns most recent commit of the given branch that is on origin/main.
func firstOriginMainCommit(branchName string) string {
	if !getLocalHasBranchOrDie(branchName) {
		panic("Branch does not exist " + branchName)
	}
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "merge-base", "origin/"+GetMainBranchOrDie(), branchName))
}

// Returns whether branchName is on remote.
func RemoteHasBranch(branchName string) bool {
	remoteBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-r", "--list", "origin/"+branchName))
	return remoteBranch != ""
}

func getLocalHasBranchOrDie(branchName string) bool {
	hasBranch, err := localHasBranch(branchName)
	if err != nil {
		panic(err)
	}
	return hasBranch
}

func localHasBranch(branchName string) (bool, error) {
	out, err := ex.Execute(ex.ExecuteOptions{}, "git", "branch", "--list", branchName)
	if err != nil {
		return false, err
	}
	localBranch := strings.TrimSpace(out)
	return localBranch != "", nil
}

func requireMainBranch() {
	if GetCurrentBranchName() != GetMainBranchOrDie() {
		panic("Must be run from " + GetMainBranchOrDie() + " branch")
	}
}

func requireCommitOnMain(commit string) {
	if commit == GetMainBranchOrDie() {
		return
	}
	newCommits := getNewCommits("HEAD")
	if !slices.ContainsFunc(newCommits, func(gitLog GitLog) bool {
		return gitLog.Commit == commit
	}) {
		panic("Commit " + commit + " does not exist on " + GetMainBranchOrDie() + ". Check `sd log` for available commits.")
	}
}

// Returns current branch name.
func GetCurrentBranchName() string {
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "HEAD"))
}

func stash(forName string) bool {
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before "+forName))
	if strings.HasPrefix(stashResult, "Saved working") {
		slog.Info(stashResult)
		return true
	}
	return false
}

func popStash(popStash bool) {
	if popStash {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "pop")
		slog.Info("Popped stash back")
	}
}
