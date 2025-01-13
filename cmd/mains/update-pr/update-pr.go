package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stacked-diff-workflow/cmd/stacked-diff"
	"strings"
)

func main() {
	var logFlags int
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			sd.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				sd.White+"Flags:"+sd.Reset+"\n")
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
		commitsToCherryPick[0] = strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	shouldPopStash := false
	stashResult := strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "stash", "save", "-u", "before update-pr "+flag.Arg(0)))
	if strings.HasPrefix(stashResult, "Saved working") {
		log.Println(stashResult)
		shouldPopStash = true
	}
	log.Println("Switching to branch", branchInfo.BranchName)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	log.Println("Fast forwarding in case there were any commits made via github web interface")
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "fetch", "origin", branchInfo.BranchName)
	forcePush := false
	if _, err := sd.Execute(sd.ExecuteOptions{}, "git", "merge", "--ff-only", "origin/"+branchInfo.BranchName); err != nil {
		log.Println("Could not fast forward to match origin. Rebasing instead.", err)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "rebase", "origin", branchInfo.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	log.Println("Cherry picking", commitsToCherryPick)
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	cherryPickOutput, cherryPickError := sd.Execute(sd.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		log.Println("First attempt at cherry-pick failed")
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseBranch := sd.FirstOriginMainCommit(sd.GetMainBranch())
		log.Println("Rebasing with the base commit on "+sd.GetMainBranch()+" branch, ", rebaseBranch,
			", in case the local "+sd.GetMainBranch()+" was rebased with origin/"+sd.GetMainBranch())
		rebaseOutput, rebaseError := sd.Execute(sd.ExecuteOptions{}, "git", "rebase", rebaseBranch)
		if rebaseError != nil {
			log.Println(sd.Red+"Could not rebase, aborting..."+sd.Reset, rebaseOutput)
			sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "rebase", "--abort")
			sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
			sd.PopStash(shouldPopStash)
			os.Exit(1)
		}
		log.Println("Cherry picking again", commitsToCherryPick)
		cherryPickOutput, cherryPickError = sd.Execute(sd.ExecuteOptions{}, "git", cherryPickArgs...)
		if cherryPickError != nil {
			log.Println(sd.Red+"Could not cherry-pick, aborting..."+sd.Reset, cherryPickArgs, cherryPickOutput, cherryPickError)
			sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "cherry-pick", "--abort")
			sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
			sd.PopStash(shouldPopStash)
			os.Exit(1)
		}
		forcePush = true
	}
	log.Println("Pushing to remote")
	if forcePush {
		if _, err := sd.Execute(sd.ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName); err != nil {
			log.Println("Regular push failed, force pushing instead.")
			sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "push", "-f", "origin", branchInfo.BranchName)
		}
	} else {
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName)
	}
	log.Println("Switching back to " + sd.GetMainBranch())
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
	log.Println("Rebasing, marking as fixup", commitsToCherryPick, "for target", branchInfo.CommitHash)
	options := sd.ExecuteOptions{EnvironmentVariables: make([]string, 1), PipeToStdout: true}
	options.EnvironmentVariables[0] = "GIT_SEQUENCE_EDITOR=sequence-editor-mark-as-fixup " + branchInfo.CommitHash + " " + strings.Join(commitsToCherryPick, " ")
	sd.ExecuteOrDie(options, "git", "rebase", "-i", branchInfo.CommitHash+"^")
	sd.PopStash(shouldPopStash)
}
