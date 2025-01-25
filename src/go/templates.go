package stackeddiff

import (
	"bytes"
	_ "embed"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	ex "stackeddiff/execute"
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
}

func GetBranchInfo(commitOrPullRequest string) BranchInfo {
	if commitOrPullRequest == "" {
		commitOrPullRequest = ex.GetMainBranch()
	}
	var info BranchInfo
	if _, err := strconv.Atoi(commitOrPullRequest); len(commitOrPullRequest) < 7 && err == nil {
		slog.Debug("Using commitOrPullRequest as a pull request number " + commitOrPullRequest)

		branchName := ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", commitOrPullRequest, "--json", "headRefName", "-q", ".headRefName")
		// Fetch the branch in case the lastest commit is only on GitHub.
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", branchName)
		prCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", commitOrPullRequest, "--json", "commits", "-q", "[.commits[].oid] | first"))
		summary := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%s", prCommit))
		thisBranchCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log", "--grep", "^"+regexp.QuoteMeta(summary)+"$", "--format=%h"))
		if thisBranchCommit == "" {
			log.Fatal("Could not find associated commit for PR (\"", summary, "\") in "+ex.GetMainBranch())
		}
		info = BranchInfo{CommitHash: thisBranchCommit, BranchName: branchName}
		slog.Info("Using pull request " + commitOrPullRequest + ", commit " + info.CommitHash + ", branch " + info.BranchName)
	} else {
		slog.Debug("Using commitOrPullRequest as a commit hash " + commitOrPullRequest)

		info = BranchInfo{CommitHash: commitOrPullRequest, BranchName: GetBranchForCommit(commitOrPullRequest)}
		slog.Info("Using commit " + info.CommitHash + ", branch " + info.BranchName)
	}
	return info

}

func GetBranchForCommit(commit string) string {
	name := runTemplate("branch-name.template", branchNameTemplateText, getBranchTemplateData(commit))
	if ex.GetMainBranch() == "master" {
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
	commitSummary := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "show", "--no-patch", "--format=%s", commitHash))
	commitBody := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "show", "--no-patch", "--format=%b", commitHash))
	commitSummaryCleaned := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%f", commitHash))
	expression := regexp.MustCompile("^(\\S+-[[:digit:]]+ )?(.*)")
	summaryMatches := expression.FindStringSubmatch(commitSummary)
	return templateData{
		Username:                   GetUsername(),
		TicketNumber:               strings.TrimSpace(summaryMatches[1]),
		CommitBody:                 commitBody,
		CommitSummary:              commitSummary,
		CommitSummaryWithoutTicket: summaryMatches[2],
		CommitSummaryCleaned:       commitSummaryCleaned,
		FeatureFlag:                featureFlag,
	}
}

func getBranchTemplateData(commitHash string) branchTemplateData {
	return branchTemplateData{
		Username:             GetUsername(),
		CommitSummaryCleaned: strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%f", commitHash)),
	}
}

func GetUsername() string {
	email := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "user.email"))
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

type GitLog struct {
	Commit  string
	Subject string
}

func newGitLogs(logsRaw string) []GitLog {
	logLines := strings.Split(strings.TrimSpace(logsRaw), "\n")
	var logs []GitLog
	for _, logLine := range logLines {
		spaceIndex := strings.Index(logLine, " ")
		if spaceIndex == -1 {
			// No git logs.
			continue
		}
		logs = append(logs, GitLog{Commit: logLine[0:spaceIndex], Subject: logLine[spaceIndex+1:]})
	}
	return logs
}

func GetAllCommits() []GitLog {
	gitArgs := []string{"--no-pager", "log", "--pretty=format:%h %s", "--abbrev-commit"}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

func GetNewCommits(compareFromRemoteBranch string, to string) []GitLog {
	gitArgs := []string{"--no-pager", "log", "--pretty=format:%h %s", "--abbrev-commit"}
	if RemoteHasBranch(compareFromRemoteBranch) {
		gitArgs = append(gitArgs, "origin/"+compareFromRemoteBranch+".."+to)
	}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

// Returns first commit of the given branch that is on origin/main, or "" if the branch is not on remote.
func FirstOriginMainCommit(branchName string) string {
	// Verify that remote has branch, there is no origin commit.
	if !RemoteHasBranch(branchName) {
		return ""
	}
	allNewCommits := GetNewCommits(branchName, "HEAD")
	return allNewCommits[len(allNewCommits)-1].Commit + "~1"
}

func RemoteHasBranch(branchName string) bool {
	remoteBranches := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-r", "--list", "origin/"+branchName))
	return remoteBranches != ""
}

func RequireMainBranch() {
	if GetCurrentBranchName() != ex.GetMainBranch() {
		log.Fatal("Must be run from " + ex.GetMainBranch() + " branch")
	}
}

func GetCurrentBranchName() string {
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "HEAD"))
}

func Stash(forName string) bool {
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before "+forName))
	if strings.HasPrefix(stashResult, "Saved working") {
		log.Println(stashResult)
		return true
	}
	return false
}

func PopStash(popStash bool) {
	if popStash {
		``
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "pop")
		log.Println("Popped stash back")
	}
}
