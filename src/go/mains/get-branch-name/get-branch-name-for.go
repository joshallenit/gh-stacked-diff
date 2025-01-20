package main

import (
	"fmt"
	"io"
	"log"
	"os"
	sd "stackeddiff"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Outputs the branch name for a given commit hash or pull request number. Useful for custom scripting.")
		fmt.Println("get-branch-name-for <commit hash or pull request number>")
		os.Exit(1)
	}
	log.SetOutput(io.Discard)
	branchName := sd.GetBranchInfo(os.Args[1]).BranchName
	fmt.Print(branchName)
}
