package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var silent bool
	flag.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Missing pullRequestNumber or commitHash")
		flag.Usage()
		os.Exit(1)
	}
	branchName := GetBranchInfo(flag.Arg(0)).BranchName
	for getMergedAt(branchName) == "" {
		log.Println("Not merged yet...")
		time.Sleep(5 * time.Minute)
	}
	log.Println("Merged!")
	if !silent {
		Execute("say", "P R has been merged")
	}
}

func getMergedAt(branchName string) string {
	return Execute("gh", "pr", "view", branchName, "--json", "mergedAt", "--jq", ".mergedAt")
}
