package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

//go:embed config/branch-name.template
var branchNameTemplateText string

//go:embed config/pr-title.template
var prTitleTemplateText string

//go:embed config/pr-description.template
var prDescriptionTemplateText string

type BranchInfo struct {
	CommitHash string
	BranchName string
}

type PullRequestText struct {
	Title       string
	Description string
}

type branchTemplateData struct {
	Username             string
	CommitSummaryCleaned string
}

type templateData struct {
	TicketNumber               string
	Username                   string
	CommitBody                 string
	CommitSummary              string
	CommitSummaryCleaned       string
	CommitSummaryWithoutTicket string
	FeatureFlag                string
	CodeOwners                 string
}

func GetBranchInfo(commitOrPullRequest string) BranchInfo {
	if commitOrPullRequest == "" {
		commitOrPullRequest = GetMainBranch()
	}
	var info BranchInfo
	if _, err := strconv.Atoi(commitOrPullRequest); len(commitOrPullRequest) < 9 && err == nil {
		// Pull request number
		branchName := Execute("gh", "pr", "view", commitOrPullRequest, "--json", "headRefName", "-q", ".headRefName")
		// Fetch the branch in case the lastest commit is only on GitHub.
		ExecuteFailable("git", "fetch", "origin", branchName)
		prCommit := Execute("gh", "pr", "view", commitOrPullRequest, "--json", "commits", "-q", "[.commits[].oid] | first")
		summary := Execute("git", "show", "--no-patch", "--format=%s", prCommit)
		thisBranchCommit := Execute("git", "log", "--grep", "^"+regexp.QuoteMeta(summary)+"$", "--format=%h")
		if thisBranchCommit == "" {
			log.Fatal("Could not find associated commit for PR (\"", summary, "\") in "+GetMainBranch())
		}
		info = BranchInfo{CommitHash: thisBranchCommit, BranchName: branchName}
		log.Print("Using pull request ", commitOrPullRequest, ", commit ", info.CommitHash, ", branch ", info.BranchName, "\n")
	} else {
		// commit hash
		info = BranchInfo{CommitHash: commitOrPullRequest, BranchName: GetBranchForCommit(commitOrPullRequest)}
		log.Print("Using commit ", info.CommitHash, ", branch ", info.BranchName, "\n")
	}
	return info

}

func GetBranchForCommit(commit string) string {
	name := runTemplate("branch-name.template", branchNameTemplateText, getBranchTemplateData(commit))
	if GetMainBranch() == "master" {
		name = strings.Replace(name, "/", "-", -1)
		name = strings.Replace(name, ".", "-", -1)
	}
	name = truncateString(name, 120)
	return name
}

func truncateString(str string, maxBytes int) string {
	for i, _ := range str {
		if i >= maxBytes {
			return str[:i]
		}
	}
	return str
}

func GetPullRequestText(commitHash string, featureFlag string) PullRequestText {
	data := getTemplateData(commitHash, featureFlag)
	title := runTemplate("pr-title.template", prTitleTemplateText, data)
	description := runTemplate("pr-description.template", prDescriptionTemplateText, data)
	return PullRequestText{Description: description, Title: title}
}

func runTemplate(configFilename string, defaultTemplateText string, data any) string {
	configFile := getConfigFile(configFilename)
	var parsed *template.Template
	var err error
	if configFile != nil {
		parsed, err = template.ParseFiles(*configFile)
		if err != nil {
			log.Fatal("Could not parse ", *configFile, err)
		}
	} else {
		parsed, err = template.New("").Parse(defaultTemplateText)
		if err != nil {
			log.Fatal("Could not parse ", defaultTemplateText, err)
		}
	}
	var output bytes.Buffer
	parsed.Execute(&output, data)
	return output.String()
}

func getTemplateData(commitHash string, featureFlag string) templateData {
	commitSummary := Execute("git", "--no-pager", "show", "--no-patch", "--format=%s", commitHash)
	commitBody := Execute("git", "--no-pager", "show", "--no-patch", "--format=%b", commitHash)
	commitSummaryCleaned := Execute("git", "show", "--no-patch", "--format=%f", commitHash)
	expression := regexp.MustCompile("^(\\S+[[:digit:]]+ )?(.*)")
	summaryMatches := expression.FindStringSubmatch(commitSummary)
	return templateData{
		Username:                   getUsername(),
		TicketNumber:               strings.TrimSpace(summaryMatches[1]),
		CommitBody:                 commitBody,
		CommitSummary:              commitSummary,
		CommitSummaryWithoutTicket: summaryMatches[2],
		CommitSummaryCleaned:       commitSummaryCleaned,
		FeatureFlag:                featureFlag,
		CodeOwners:                 ChangedFilesOwnersString(true),
	}
}

func getBranchTemplateData(commitHash string) branchTemplateData {
	return branchTemplateData{
		Username:             getUsername(),
		CommitSummaryCleaned: Execute("git", "show", "--no-patch", "--format=%f", commitHash),
	}
}

func getUsername() string {
	email := Execute("git", "config", "user.email")
	return email[0:strings.Index(email, "@")]
}

func getConfigFile(filenameWithoutPath string) *string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not get home dir", err)
	}
	fullPath := home + "/.stacked-diff-workflow/" + filenameWithoutPath
	if _, err := os.Stat(fullPath); err == nil {
		return &fullPath
	} else {
		return nil
	}
}

// Returns first commit of the given branch that is on origin/main.
func FirstOriginMainCommit(branchName string) string {
	allNewCommits := strings.Fields(Execute("git", "--no-pager", "log", "origin/"+GetMainBranch()+".."+branchName, "--pretty=format:%h", "--abbrev-commit"))
	if len(allNewCommits) == 0 {
		log.Fatal("No commits on ", branchName, "other than what is on "+GetMainBranch()+", nothing to create a commit from")
	}
	return allNewCommits[len(allNewCommits)-1] + "~1"
}

func RequireMainBranch() {
	if GetCurrentBranchName() != GetMainBranch() {
		log.Fatal("Must be run from " + GetMainBranch() + " branch")
	}
}

func GetCurrentBranchName() string {
	return Execute("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func PopStash(popStash bool) {
	if popStash {
		Execute("git", "stash", "pop")
		log.Println("Popped stash back")
	}
}
