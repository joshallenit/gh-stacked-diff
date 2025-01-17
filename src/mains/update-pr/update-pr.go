package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	ex "stacked-diff-workflow/src/execute"
	sd "stacked-diff-workflow/src/stacked-diff"
	"strings"
)

func main() {
	var logFlags int
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			ex.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(logFlags)
	sd.RequireMainBranch()
	branchInfo := sd.GetBranchInfo(flag.Arg(0))
	var commitsToCherryPick []string
	if len(flag.Args()) > 1 {
		commitsToCherryPick = flag.Args()[1:]
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	shouldPopStash := false
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before update-pr "+flag.Arg(0)))
	if strings.HasPrefix(stashResult, "Saved working") {
		log.Println(stashResult)
		shouldPopStash = true
	}
	log.Println("Switching to branch", branchInfo.BranchName)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	log.Println("Fast forwarding in case there were any commits made via github web interface")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", branchInfo.BranchName)
	forcePush := false
	if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "merge", "--ff-only", "origin/"+branchInfo.BranchName); err != nil {
		log.Println("Could not fast forward to match origin. Rebasing instead.", err)
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "origin", branchInfo.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	log.Println("Cherry picking", commitsToCherryPick)
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	cherryPickOutput, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		log.Println("First attempt at cherry-pick failed")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseBranch := sd.FirstOriginMainCommit(ex.GetMainBranch())
		log.Println("Rebasing with the base commit on "+ex.GetMainBranch()+" branch, ", rebaseBranch,
			", in case the local "+ex.GetMainBranch()+" was rebased with origin/"+ex.GetMainBranch())
		rebaseOutput, rebaseError := ex.Execute(ex.ExecuteOptions{}, "git", "rebase", rebaseBranch)
		if rebaseError != nil {
			log.Println(ex.Red+"Could not rebase, aborting..."+ex.Reset, rebaseOutput)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			sd.PopStash(shouldPopStash)
			os.Exit(1)
		}
		log.Println("Cherry picking again", commitsToCherryPick)
		cherryPickOutput, cherryPickError = ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
		if cherryPickError != nil {
			log.Println(ex.Red+"Could not cherry-pick, aborting..."+ex.Reset, cherryPickArgs, cherryPickOutput, cherryPickError)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			sd.PopStash(shouldPopStash)
			os.Exit(1)
		}
		forcePush = true
	}
	log.Println("Pushing to remote")
	if forcePush {
		if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName); err != nil {
			log.Println("Regular push failed, force pushing instead.")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "-f", "origin", branchInfo.BranchName)
		}
	} else {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName)
	}
	log.Println("Switching back to " + ex.GetMainBranch())
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
	log.Println("Rebasing, marking as fixup", commitsToCherryPick, "for target", branchInfo.CommitHash)
	options := ex.ExecuteOptions{EnvironmentVariables: make([]string, 1), Output: ex.NewStandardOutput()}
	options.EnvironmentVariables[0] = "GIT_SEQUENCE_EDITOR=sequence-editor-mark-as-fixup " + branchInfo.CommitHash + " " + strings.Join(commitsToCherryPick, " ")
	ex.ExecuteOrDie(options, "git", "rebase", "-i", branchInfo.CommitHash+"^")
	sd.PopStash(shouldPopStash)
}
