package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Next `gh pr view 83824 --json latestReviews` and ensure developer is already not approved so that the review is not dismissed
type PullRequestChecksStatus struct {
	Pending int
	Failing int
	Passing int
	Total   int
}

func main() {
	var reviewers string
	flag.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers. "+
		"Falls back to "+White+"PR_REVIEWERS"+Reset+" environment variable. "+
		"You can specify more than one reviewer using a comma-delimited string.")
	var whenChecksPass bool
	var pollFrequency time.Duration
	var defaultPollFrequency time.Duration = 5 * time.Minute
	var silent bool
	var minChecks int
	flag.BoolVar(&whenChecksPass, "when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	flag.DurationVar(&pollFrequency, "poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	flag.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flag.IntVar(&minChecks, "min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			Reset+"Mark a Draft PR as \"Ready for Review\" and automatically add reviewers.\n"+
				"\n"+
				"add-reviewers [flags] <commit hash or pull request number>\n"+
				"\n"+
				White+"Flags:"+Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if reviewers == "" {
		reviewers = os.Getenv("PR_REVIEWERS")
		if reviewers == "" {
			log.Fatal("Use reviewers flag or set PR_REVIEWERS environment variable")
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < flag.NArg(); i++ {
		wg.Add(1)
		go checkBranch(&wg, flag.Arg(i), whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}
	wg.Wait()
}

func checkBranch(wg *sync.WaitGroup, commitOrPullRequest string, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	branchName := GetBranchInfo(commitOrPullRequest).BranchName
	if whenChecksPass {
		for {
			summary := getChecksStatus(branchName)
			if summary.Failing > 0 {
				if !silent {
					Execute("say", "Checks failed")
				}
				log.Print("Checks failed for ", commitOrPullRequest, ". "+
					"Total: ", summary.Total,
					" | Passed: ", summary.Passing,
					" | Pending: ", summary.Pending,
					" | Failed: ", summary.Failing,
					"\n")
				os.Exit(1)
			}

			if summary.Total < minChecks {
				log.Println("Waiting for at least", minChecks, "checks to be added to PR. Currently only ", summary.Total)
			} else if summary.Passing == summary.Total {
				log.Println("All", summary.Total, "checks passed")
				break
			} else if summary.Passing == 0 {
				log.Print("Checks pending for ", commitOrPullRequest, ". Completed: 0%\n")
			} else {
				log.Print("Checks pending for ", commitOrPullRequest, ". Completed: ", int32(float32(summary.Passing)/float32(summary.Total)*100), "%\n")
			}
			time.Sleep(pollFrequency)
		}
	}
	log.Println("Marking PR as ready for review")
	Execute("gh", "pr", "ready", branchName)
	log.Println("Waiting 10 seconds for any automatically assigned reviewers to be added...")
	time.Sleep(10 * time.Second)
	prUrl := Execute("gh", "pr", "edit", branchName, "--add-reviewer", reviewers)
	log.Println("Added reviewers", reviewers, "to", prUrl)
	wg.Done()
}

/*
 * Logic copied from https://github.com/cli/cli/blob/57fbe4f317ca7d0849eeeedb16c1abc21a81913b/api/queries_pr.go#L258-L274
 */
func getChecksStatus(branchName string) PullRequestChecksStatus {
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
