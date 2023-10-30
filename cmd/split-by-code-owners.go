package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var slackToGithubTeamMap map[string]string
var changedFiles []string

func main() {
	slackToGithubTeamMap = make(map[string]string)
	slackToGithubTeamMap["Expansion - Activation"] = "activation"
	slackToGithubTeamMap["UI Infrastructure"] = "ui-infra"
	slackToGithubTeamMap["Android Application Infrastructure"] = "infra"
	slackToGithubTeamMap["Platform - App Interactivity"] = "platform"
	slackToGithubTeamMap["Enterprise - Mobile"] = "enterprise"
	// There is no Android files team at the moment.
	slackToGithubTeamMap["Files"] = "csc"
	var createPrs bool
	var shouldCreateBranches bool
	flag.BoolVar(&createPrs, "create-prs", true, "Create pull requests for each branch")
	flag.BoolVar(&shouldCreateBranches, "create-branches", true, "Create branches")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			Reset+"Split a commit by code owners.\n"+
				"split-by-code-owners <commit hash>\n")
		flag.PrintDefaults()
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	changedFiles = GetChangedFiles(flag.Arg(0))
	deleteFromSlice(changedFiles, "\"\"")
	if len(changedFiles) != 0 {
		log.Println("Splitting up", len(changedFiles), "files")
		branchInfo := GetBranchInfo(flag.Arg(0))

		branches := createBranches(branchInfo, true, shouldCreateBranches)
		branches = append(branches, createBranches(branchInfo, false, shouldCreateBranches)...)
		if len(changedFiles) != 0 {
			log.Fatal("Impossible, not all files included:", len(changedFiles), changedFiles)
		}
		if createPrs {
			for _, branchName := range branches {
				log.Println("Creating PR for", branchName)
				Execute("git", "switch", branchName)
				commitHash := Execute("git", "rev-parse", "HEAD")
				prText := GetPullRequestText(commitHash, "")
				// Sleep to avoid github crapping out with kex_exchange_identification or LFS lock.
				time.Sleep(10 * time.Second)
				Execute("git", "-c", "push.default=current", "push", "-f")
				time.Sleep(10 * time.Second)
				if url, err := ExecuteFailable("gh", "pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--draft"); err != nil {
					log.Println("Could not create PR", err)
				} else {
					log.Println("Created PR", url)
				}
			}
		}
	}
}

func createBranches(branchInfo BranchInfo, useGithub bool, shouldCreateBranches bool) []string {
	originalBranch := GetCurrentBranchName()
	var branches []string
	for ownerName, filenames := range ChangedFilesOwners(useGithub, changedFiles) {
		if useGithub && ownerName == "unowned" {
			continue
		}
		shortTeamName := gitTeamToShortName(ownerName)
		branchName := branchInfo.BranchName + "-for-" + shortTeamName
		branches = append(branches, branchName)
		if shouldCreateBranches {
			log.Println("Creating branch", branchName, "with", len(filenames), "files")
			if _, err := ExecuteFailable("git", "checkout", "-b", branchName); err != nil {
				log.Println("Adding to existing branch")
				Execute("git", "checkout", branchName)
				Execute("git", "reset", "--hard", "head")
			} else {
				Execute("git", "reset", "--hard", "origin/"+GetMainBranch())
			}
			diff := ExecuteWithOptions(ExecuteOptions{TrimSpace: false}, "git", "diff", "--binary", branchInfo.CommitHash+"~", branchInfo.CommitHash)
			ExecuteWithOptions(ExecuteOptions{Stdin: &diff, AbortOnFailure: false}, "git", "apply", "--reject")
		}
		gitAddArgs := []string{"add"}
		for _, filename := range filenames {
			changedFiles = deleteFromSlice(changedFiles, filename)
			gitAddArgs = append(gitAddArgs, filename)
		}
		if shouldCreateBranches {
			Execute("git", gitAddArgs...)
			summary := Execute("git", "show", "--no-patch", "--format=%s", branchInfo.CommitHash)
			Execute("git", "commit", "-m", summary+" for "+shortTeamName)
			Execute("git", "reset", "--hard", "HEAD")
			Execute("git", "clean", "--force")
		}
	}
	if shouldCreateBranches {
		log.Println("Switching back to original branch")
		Execute("git", "switch", originalBranch)
	}
	return branches
}

func gitTeamToShortName(gitTeamName string) string {
	// @tinyspeck/android-commons-codeowners to commons
	mapped := slackToGithubTeamMap[gitTeamName]
	if mapped != "" {
		return mapped
	}
	shortName := strings.Replace(gitTeamName, "@tinyspeck/android-", "", -1)
	shortName = strings.Replace(shortName, "-codeowners", "", -1)
	shortName = strings.Replace(shortName, "&", "and", -1)
	shortName = strings.Replace(shortName, ",", "-and-", -1)
	shortName = strings.Replace(shortName, " - ", "-", -1)
	shortName = strings.Replace(shortName, " ", "-", -1)
	shortName = strings.ToLower(shortName)
	return shortName
}

func deleteFromSlice(slice []string, elem string) []string {
	for i, next := range slice {
		if next == elem {
			modified := slice[0:i]
			if i+1 < len(slice) {
				modified = append(modified, slice[i+1:len(slice)]...)
			}
			return modified
		}
	}
	return slice
}
