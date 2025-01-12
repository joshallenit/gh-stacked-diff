package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	var logFlags int
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				White+"Flags:"+Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(logFlags)
	RequireMainBranch()
	branchInfo := GetBranchInfo(flag.Arg(0))
	var commitsToCherryPick []string
	if len(flag.Args()) > 1 {
		commitsToCherryPick = flag.Args()[1:]
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	shouldPopStash := false
	stashResult := strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "stash", "save", "-u", "before update-pr "+flag.Arg(0)))
	if strings.HasPrefix(stashResult, "Saved working") {
		log.Println(stashResult)
		shouldPopStash = true
	}
	log.Println("Switching to branch", branchInfo.BranchName)
	ExecuteOrDie(ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	log.Println("Fast forwarding in case there were any commits made via github web interface")
	ExecuteOrDie(ExecuteOptions{}, "git", "fetch", "origin", branchInfo.BranchName)
	forcePush := false
	if _, err := Execute(ExecuteOptions{}, "git", "merge", "--ff-only", "origin/"+branchInfo.BranchName); err != nil {
		log.Println("Could not fast forward to match origin. Rebasing instead.", err)
		ExecuteOrDie(ExecuteOptions{}, "git", "rebase", "origin", branchInfo.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	log.Println("Cherry picking", commitsToCherryPick)
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	cherryPickOutput, cherryPickError := Execute(ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		log.Println("First attempt at cherry-pick failed")
		ExecuteOrDie(ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseBranch := FirstOriginMainCommit(GetMainBranch())
		log.Println("Rebasing with the base commit on "+GetMainBranch()+" branch, ", rebaseBranch,
			", in case the local "+GetMainBranch()+" was rebased with origin/"+GetMainBranch())
		rebaseOutput, rebaseError := Execute(ExecuteOptions{}, "git", "rebase", rebaseBranch)
		if rebaseError != nil {
			log.Println(Red+"Could not rebase, aborting..."+Reset, rebaseOutput)
			ExecuteOrDie(ExecuteOptions{}, "git", "rebase", "--abort")
			ExecuteOrDie(ExecuteOptions{}, "git", "switch", GetMainBranch())
			PopStash(shouldPopStash)
			os.Exit(1)
		}
		log.Println("Cherry picking again", commitsToCherryPick)
		cherryPickOutput, cherryPickError = Execute(ExecuteOptions{}, "git", cherryPickArgs...)
		if cherryPickError != nil {
			log.Println(Red+"Could not cherry-pick, aborting..."+Reset, cherryPickArgs, cherryPickOutput, cherryPickError)
			ExecuteOrDie(ExecuteOptions{}, "git", "cherry-pick", "--abort")
			ExecuteOrDie(ExecuteOptions{}, "git", "switch", GetMainBranch())
			PopStash(shouldPopStash)
			os.Exit(1)
		}
		forcePush = true
	}
	log.Println("Pushing to remote")
	if forcePush {
		if _, err := Execute(ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName); err != nil {
			log.Println("Regular push failed, force pushing instead.")
			ExecuteOrDie(ExecuteOptions{}, "git", "push", "-f", "origin", branchInfo.BranchName)
		}
	} else {
		ExecuteOrDie(ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName)
	}
	log.Println("Switching back to " + GetMainBranch())
	ExecuteOrDie(ExecuteOptions{}, "git", "switch", GetMainBranch())
	log.Println("Rebasing, marking as fixup", commitsToCherryPick, "for target", branchInfo.CommitHash)
	options := ExecuteOptions{EnvironmentVariables: make([]string, 1), PipeToStdout: true}
	options.EnvironmentVariables[0] = "GIT_SEQUENCE_EDITOR=sequence-editor-mark-as-fixup " + branchInfo.CommitHash + " " + strings.Join(commitsToCherryPick, " ")
	ExecuteOrDie(options, "git", "rebase", "-i", branchInfo.CommitHash+"^")
	PopStash(shouldPopStash)
}
