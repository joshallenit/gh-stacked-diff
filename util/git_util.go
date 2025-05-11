package util

import (
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// Cached value of main branch name.
var mainBranchNameForHelp string

var mainBranchNameFromGitLog string

// Cached value of user email.
var userEmail string

// Cached repository name.
var repoName string
var repoNameOnce *sync.Once = new(sync.Once)

// Cached repository name with owner.
var repoNameWithOwner string
var repoNameWithOwnerOnce *sync.Once = new(sync.Once)

// Cached logged in username
var loggedInUsername string
var loggedInUsernameOnce *sync.Once = new(sync.Once)

// Returns name of main branch, or panics if cannot be determined.
func GetMainBranchOrDie() string {
	out, err := getMainBranchFromGitLog()
	if err == nil {
		return out
	}
	out, err = Execute(ExecuteOptions{}, "git", "rev-parse")
	if err != nil {
		panic("Not in a git repository. Must be run from a git repository.\n" + out + ": " + err.Error())
	}

	out, err = Execute(ExecuteOptions{}, "git", "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		panic("Remote repository is empty.\n" +
			"Push an initial inconsequential commit to origin/main and try again. \n" +
			"Using a repository without an initial remote commit is not recommended because git \n" +
			"requires special handling for the root commit, and you might accidentially \n" +
			"create more than one root commit.\n" + out + ": " + err.Error())
	}

	setRemoteHead()
	out, err = getMainBranchFromGitLog()
	if err != nil {
		panic("Remote repository not setup.\n" + out + ": " + err.Error())
	}
	return out
}

// Returns name of main branch, or "main" if cannot be determined. For use by CLI help.
func GetMainBranchForHelp() string {
	if mainBranchNameForHelp != "" {
		return mainBranchNameForHelp
	}
	mainBranch, err := getMainBranchFromGitLog()
	if err != nil {
		mainBranchNameForHelp = "main"
	} else {
		mainBranchNameForHelp = mainBranch
	}
	return mainBranchNameForHelp
}

func getMainBranchFromGitLog() (string, error) {
	if mainBranchNameFromGitLog != "" {
		return mainBranchNameFromGitLog, nil
	}
	remoteMainBranch, err := Execute(ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "origin/HEAD")
	if err != nil {
		return remoteMainBranch, err
	}
	remoteMainBranch = strings.TrimSpace(remoteMainBranch)
	mainBranchNameFromGitLog = remoteMainBranch[strings.Index(remoteMainBranch, "/")+1:]
	return mainBranchNameFromGitLog, nil
}

func setRemoteHead() {
	currentBranch := GetCurrentBranchName()
	defaultBranch, err := Execute(ExecuteOptions{}, "git", "config", "init.defaultBranch")
	if err != nil {
		// git config init.defaultBranch will fail if default branch is not setup.
		defaultBranch = "master"
	} else {
		defaultBranch = strings.TrimSpace(defaultBranch)
	}
	if currentBranch == defaultBranch || currentBranch == "main" {
		slog.Warn("Setting remote head to " + currentBranch + " because it is not set.")
		out, err := Execute(ExecuteOptions{}, "git", "remote", "set-head", "origin", currentBranch)
		if err != nil {
			panic("Remote repository not setup.\n" + out)
		}
	} else {
		panic("Remote head is not set, and it cannot be set automatically because current branch is not default (" + defaultBranch + ") or main.")
	}
}

func GetUsername() string {
	if userEmail == "" {
		userEmailRaw := strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "config", "user.email"))
		userEmail = userEmailRaw[0:strings.Index(userEmailRaw, "@")]
	}
	return userEmail
}

// Returns most recent commit of the given branch that is on origin/main.
func FirstOriginMainCommit(branchName string) string {
	if !GetLocalHasBranchOrDie(branchName) {
		panic("Branch does not exist " + branchName)
	}
	return strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "merge-base", "origin/"+GetMainBranchOrDie(), branchName))
}

// Returns whether branchName is on remote.
func RemoteHasBranch(branchName string) bool {
	remoteBranch := strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "branch", "-r", "--list", "origin/"+branchName))
	return remoteBranch != ""
}

func GetLocalHasBranchOrDie(branchName string) bool {
	hasBranch, err := localHasBranch(branchName)
	if err != nil {
		panic(err)
	}
	return hasBranch
}

func localHasBranch(branchName string) (bool, error) {
	out, err := Execute(ExecuteOptions{}, "git", "branch", "--list", branchName)
	if err != nil {
		return false, err
	}
	localBranch := strings.TrimSpace(out)
	return localBranch != "", nil
}

func RequireMainBranch() {
	if GetCurrentBranchName() != GetMainBranchOrDie() {
		panic("Must be run from " + GetMainBranchOrDie() + " branch")
	}
}

// Returns current branch name.
func GetCurrentBranchName() string {
	return strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "HEAD"))
}

func Stash(forName string) bool {
	stashResult := strings.Split(strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "stash", "save", "-u", "before "+forName)), "\n")
	if len(stashResult) > 0 && strings.HasPrefix(stashResult[len(stashResult)-1], "Saved working") {
		slog.Info(stashResult[len(stashResult)-1])
		return true
	}
	return false
}

func PopStash(popStash bool) {
	if popStash {
		ExecuteOrDie(ExecuteOptions{}, "git", "stash", "pop")
		slog.Info("Popped stash back")
	}
}

// Returns "repository-owner/repository-name".
func GetRepoNameWithOwner() string {
	if repoNameWithOwner == "" {
		repoNameWithOwnerOnce.Do(func() {
			out := ExecuteOrDie(ExecuteOptions{},
				"gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
			repoNameWithOwner = strings.TrimSpace(out)
		})
	}
	return repoNameWithOwner
}

func GetRepoName() string {
	if repoName == "" {
		repoNameOnce.Do(func() {
			out := ExecuteOrDie(ExecuteOptions{},
				"git", "rev-parse", "--show-toplevel")
			_, repoName = filepath.Split(strings.TrimSpace(out))
		})
	}
	return repoName
}

func GetLoggedInUsername() string {
	if loggedInUsername == "" {
		loggedInUsernameOnce.Do(func() {
			out := ExecuteOrDie(ExecuteOptions{},
				"gh", "api", "https://api.github.com/user", "--jq", ".login")
			loggedInUsername = strings.TrimSpace(out)
		})
	}
	return loggedInUsername
}

/*
Returns users that have already approved latest commit.

Example output of gh pr view:

$ gh pr view mybranch --json "reviews"

	{
	  "reviews": [
	    {
	      "id": "PRR_kwDODeVIac6f37Qq",
	      "author": {
	        "login": "mybestie"
	      },
	      "authorAssociation": "MEMBER",
	      "body": "",
	      "submittedAt": "2025-03-13T14:47:31Z",
	      "includesCreatedEdit": false,
	      "reactionGroups": [],
	      "state": "COMMENTED",
	      "commit": {
	        "oid": "af01bdf8eb5649956096a608717f7de5eeb97e45"
	      }
	    },
	    {
	      "id": "PRR_kwDODeVIac6f5jeG",
	      "author": {
	        "login": "myfave"
	      },
	      "authorAssociation": "MEMBER",
	      "body": "",
	      "submittedAt": "2025-03-13T16:32:44Z",
	      "includesCreatedEdit": false,
	      "reactionGroups": [],
	      "state": "APPROVED",
	      "commit": {
	        "oid": "af01bdf8eb5649956096a608717f7de5eeb97e45"
	      }
	    }
	  ]
	}
*/
func GetAllApprovingUsers(branchName string) []string {
	lastCommit := GetBranchLatestCommit(branchName)
	jq := ".reviews[] | select(.state == \"APPROVED\" and .commit.oid == \"" + lastCommit + "\") | .author.login"
	out := ExecuteOrDie(ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "reviews", "--jq", jq)
	approvingUsers := strings.Fields(out)
	slices.Sort(approvingUsers)
	return slices.Compact(approvingUsers)
}

// Returns full commit hash of branch with name of branchName, or "" if no such branch.
func GetBranchLatestCommit(branchName string) string {
	out, err := Execute(ExecuteOptions{}, "git", "log", "-n", "1", "--pretty=format:%H", branchName)
	if err != nil {
		return ""
	} else {
		return strings.TrimSpace(out)
	}
}
