package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stackeddiff"
	"strings"
	"time"

	ex "stackeddiff/execute"
)

func main() {
	var silent bool
	flag.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			ex.Reset+"Waits for a pull request to be merged. Polls PR every 5 minutes. Useful for custom scripting.\n"+
				"\n"+
				"wait-for-merge [flags] <commit hash or pull request number>\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	branchName := sd.GetBranchInfo(flag.Arg(0)).BranchName
	for getMergedAt(branchName) == "" {
		log.Println("Not merged yet...")
		time.Sleep(30 * time.Second)
	}
	log.Println("Merged!")
	if !silent {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "say", "P R has been merged")
	}
}

func getMergedAt(branchName string) string {
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "mergedAt", "--jq", ".mergedAt"))
}
