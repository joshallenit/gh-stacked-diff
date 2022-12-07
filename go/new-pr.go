package main

import (
	"log"
	"os"
)

func main() {
	branchInfo := GetBranchInfo(os.Args[1])
	prText := GetPullRequestText(branchInfo.CommitHash)
	ExecuteFailable("git", "branch", "--no-track", branchInfo.BranchName, "origin/main")
	Execute("git", "switch", branchInfo.BranchName)
	_, cherryPickError := ExecuteFailable("git", "cherry-pick", branchInfo.CommitHash)
	if cherryPickError != nil {
		Execute("git", "cherry-pick", "--abort")
		Execute("git", "switch", "main")
		return
	}
	Execute("git", "-c", "push.default=current", "push", "-f")
	Execute("gh", "pr", "create" /*"--draft", */, "--title", prText.Title, "--body", prText.Description, "--fill")
	Execute("gh", "pr", "view", "--web")
	Execute("git", "switch", "main")
}
