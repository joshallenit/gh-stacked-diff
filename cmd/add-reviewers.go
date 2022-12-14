package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type PullRequestChecksStatus struct {
	Pending int
	Failing int
	Passing int
	Total   int
}

func main() {
	var reviewers string
	flag.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers. Falls back to PR_REVIEWERS environment variable")
	var whenChecksPass bool
	var pollFrequency time.Duration
	var defaultPollFrequency time.Duration = 5 * time.Minute
	var silent bool
	flag.BoolVar(&whenChecksPass, "when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	flag.DurationVar(&pollFrequency, "poll-frequency", defaultPollFrequency, "Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	flag.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Missing pullRequestNumber or commitHash")
		flag.Usage()
		os.Exit(1)
	}
	branchName := GetBranchInfo(flag.Arg(0)).BranchName
	if reviewers == "" {
		reviewers = os.Getenv("PR_REVIEWERS")
		if reviewers == "" {
			log.Fatal("Use reviewers flag or set PR_REVIEWERS environment variable")
		}
	}
	if whenChecksPass {
		for {
			summary := getChecksStatus(branchName)
			if summary.Passing == summary.Total {
				log.Println("All", summary.Total, "checks passed")
				break
			}
			if summary.Failing > 0 {
				if !silent {
					Execute("say", "Checks failed")
				}
				log.Println("Checks failed. Total: ", summary.Total, "| Passed: ", summary.Passing, "| Pending: ", summary.Pending, "| Failed: ", summary.Failing)
				os.Exit(1)
			}
			if summary.Passing == 0 {
				log.Println("Checks pending. Completed: 0%", summary.Passing)
			} else {
				log.Println("Checks pending. Completed:", int32(float32(summary.Passing)/float32(summary.Total)*100), "%")
			}
			time.Sleep(pollFrequency)
		}
	}
	Execute("gh", "pr", "ready", branchName)
	prUrl := Execute("gh", "pr", "edit", branchName, "--add-reviewer", reviewers)
	log.Println("Added reviewers", reviewers, "to", prUrl)
}

/*
 * Logic copied from https://github.com/cli/cli/blob/57fbe4f317ca7d0849eeeedb16c1abc21a81913b/api/queries_pr.go#L258-L274
 */
func getChecksStatus(branchName string) PullRequestChecksStatus {
	// jq  ~/Downloads/test.json
	var summary PullRequestChecksStatus
	stateString := ExecuteWithOptions(ExecuteOptions{TrimSpace: false}, "gh", "pr", "view", branchName, "--json", "statusCheckRollup", "--jq", ".statusCheckRollup[] | .status, .conclusion, .state")
	scanner := bufio.NewScanner(strings.NewReader(stateString))
	for scanner.Scan() {
		status := scanner.Text()
		scanner.Scan()
		conclusion := scanner.Text()
		scanner.Scan()
		state := scanner.Text()
		if state == "" {
			if status == "COMPLETED" {
				state = conclusion
			} else {
				state = status
			}
		}
		switch state {
		case "SUCCESS", "NEUTRAL", "SKIPPED":
			summary.Passing++
		case "ERROR", "FAILURE", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED":
			summary.Failing++
		default: // "EXPECTED", "REQUESTED", "WAITING", "QUEUED", "PENDING", "IN_PROGRESS", "STALE"
			summary.Pending++
		}
		summary.Total++
	}
	return summary
}
