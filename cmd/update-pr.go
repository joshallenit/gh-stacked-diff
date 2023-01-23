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
		commitsToCherryPick[0] = Execute("git", "rev-parse", "--short", "HEAD")
	}
	log.Println("Switching to branch", branchInfo.BranchName)
	Execute("git", "switch", branchInfo.BranchName)
	log.Println("Fast forwarding in case there were any commits made via github web interface")
	Execute("git", "fetch", "origin", branchInfo.BranchName)
	forcePush := false
	if _, err := ExecuteFailable("git", "merge", "--ff-only", "origin/"+branchInfo.BranchName); err != nil {
		log.Println("Could not fast forward to match origin. Rebasing instead.", err)
		Execute("git", "rebase", "origin", branchInfo.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}
	log.Println("Cherry picking", commitsToCherryPick)
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	cherryPickOutput, cherryPickError := ExecuteFailable("git", cherryPickArgs...)
	if cherryPickError != nil {
		log.Println("Could not cherry-pick, aborting...", cherryPickArgs, cherryPickOutput, cherryPickError)
		Execute("git", "cherry-pick", "--abort")
		Execute("git", "switch", "main")
		os.Exit(1)
	}
	log.Println("Pushing to remote")
	if forcePush {
		if _, err := ExecuteFailable("git", "push", "origin", branchInfo.BranchName); err != nil {
			log.Println("Regular push failed, force pushing.", err)
			Execute("git", "push", "-f", "origin", branchInfo.BranchName)
		}
	} else {
		Execute("git", "push", "origin", branchInfo.BranchName)
	}
	log.Println("Switching back to main")
	Execute("git", "switch", "main")
	log.Println("Rebasing, marking as fixup", commitsToCherryPick, "for target", branchInfo.CommitHash)
	options := ExecuteOptions{EnvironmentVariables: make([]string, 1)}
	options.EnvironmentVariables[0] = "GIT_SEQUENCE_EDITOR=sequence-editor-mark-as-fixup " + branchInfo.CommitHash + " " + strings.Join(commitsToCherryPick, " ")
	ExecuteWithOptions(options, "git", "rebase", "-i", branchInfo.CommitHash+"^")
}
