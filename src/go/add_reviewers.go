package stackeddiff

import (
	"bufio"
	"fmt"
	"log/slog"
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

func AddReviewersToPr(commitIndicators []string, indicatorType IndicatorType, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	if reviewers == "" {
		panic("Reviewers cannot be empty")
	}
	var wg sync.WaitGroup
	for _, commitIndicator := range commitIndicators {
		wg.Add(1)
		go checkBranch(&wg, commitIndicator, indicatorType, whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}
	wg.Wait()
}

func checkBranch(wg *sync.WaitGroup, commitIndicator string, indicatorType IndicatorType, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	branchName := GetBranchInfo(commitIndicator, indicatorType).BranchName
	if whenChecksPass {
		for {
			summary := getChecksStatus(branchName)
			if summary.Failing > 0 {
				if !silent {
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "say", "Checks failed")
				}
				slog.Error(fmt.Sprint("Checks failed for ", commitIndicator, ". "+
					"Total: ", summary.Total,
					" | Passed: ", summary.Passing,
					" | Pending: ", summary.Pending,
					" | Failed: ", summary.Failing,
					"\n"))
				os.Exit(1)
			}

			if summary.Total < minChecks {
				slog.Info(fmt.Sprint("Waiting for at least", minChecks, "checks to be added to PR. Currently only ", summary.Total))
			} else if summary.Passing == summary.Total {
				slog.Info(fmt.Sprint("All", summary.Total, "checks passed"))
				break
			} else if summary.Passing == 0 {
				slog.Info(fmt.Sprint("Checks pending for ", commitIndicator, ". Completed: 0%"))
			} else {
				slog.Info(fmt.Sprint("Checks pending for ", commitIndicator, ". Completed: ", int32(float32(summary.Passing)/float32(summary.Total)*100), "%"))
			}
			time.Sleep(pollFrequency)
		}
	}
	slog.Info("Marking PR as ready for review")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "ready", branchName)
	slog.Info("Waiting 10 seconds for any automatically assigned reviewers to be added...")
	time.Sleep(10 * time.Second)
	prUrl := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "edit", branchName, "--add-reviewer", reviewers))
	slog.Info(fmt.Sprint("Added reviewers", reviewers, "to", prUrl))
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
