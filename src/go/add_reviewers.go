package stackeddiff

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	ex "stackeddiff/execute"
)

// Next `gh pr view 83824 --json latestReviews` and ensure developer is already not approved so that the review is not dismissed
type pullRequestChecksStatus struct {
	Pending int
	Failing int
	Passing int
	Total   int
}

func AddReviewersToPr(commitOrPullRequests []string, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	if reviewers == "" {
		reviewers = os.Getenv("PR_REVIEWERS")
		if reviewers == "" {
			log.Fatal("Use reviewers flag or set PR_REVIEWERS environment variable")
		}
	}
	var wg sync.WaitGroup
	for _, commitOrPullRequest := range commitOrPullRequests {
		wg.Add(1)
		go checkBranch(&wg, commitOrPullRequest, whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}
	wg.Wait()
}

func checkBranch(wg *sync.WaitGroup, commitOrPullRequest string, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	branchName := GetBranchInfo(commitOrPullRequest, IndicatorTypeGuess).BranchName
	if whenChecksPass {
		for {
			summary := getChecksStatus(branchName)
			if summary.Failing > 0 {
				if !silent {
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "say", "Checks failed")
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
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "ready", branchName)
	log.Println("Waiting 10 seconds for any automatically assigned reviewers to be added...")
	time.Sleep(10 * time.Second)
	prUrl := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "edit", branchName, "--add-reviewer", reviewers))
	log.Println("Added reviewers", reviewers, "to", prUrl)
	wg.Done()
}

/*
 * Logic copied from https://github.com/cli/cli/blob/57fbe4f317ca7d0849eeeedb16c1abc21a81913b/api/queries_pr.go#L258-L274
 */
func getChecksStatus(branchName string) pullRequestChecksStatus {
	var summary pullRequestChecksStatus
	stateString := ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "statusCheckRollup", "--jq", ".statusCheckRollup[] | .status, .conclusion, .state")
	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(stateString)))
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
