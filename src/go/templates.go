package stackeddiff

import (
	"bytes"
	"fmt"
	"slices"

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

// Delimter for git log format when a space cannot be used.
const formatDelimiter = "|stackeddiff-delim|"

//go:embed config/branch-name.template
var branchNameTemplateText string

//go:embed config/pr-title.template
var prTitleTemplateText string

//go:embed config/pr-description.template
var prDescriptionTemplateText string

// Cached value of user email.
var userEmail string

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

type IndicatorType string

const (
	IndicatorTypeCommit IndicatorType = "commit"
	IndicatorTypePr     IndicatorType = "pr"
	IndicatorTypeList   IndicatorType = "list"
	IndicatorTypeGuess  IndicatorType = "guess"
)

func (indicator IndicatorType) IsValid() bool {
	switch indicator {
	case IndicatorTypeCommit, IndicatorTypePr, IndicatorTypeList, IndicatorTypeGuess:
		return true
	default:
		return false
	}
}

func GetBranchInfo(commitIndicator string, indicatorType IndicatorType) BranchInfo {
	if !indicatorType.IsValid() {
		panic("Invalid IndicatorType " + string(indicatorType))
	}
	if commitIndicator == "" {
		commitIndicator = ex.GetMainBranch()
		indicatorType = IndicatorTypeCommit
	}
	if indicatorType == IndicatorTypeGuess {
		indicatorType = guessIndicatorType(commitIndicator)
	}

	var info BranchInfo
	switch indicatorType {
	case IndicatorTypePr:
		slog.Debug("Using commitIndicator as a pull request number " + commitIndicator)

		branchName := ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", commitIndicator, "--json", "headRefName", "-q", ".headRefName")
		// Fetch the branch in case the lastest commit is only on GitHub.
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", branchName)
		prCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", commitIndicator, "--json", "commits", "-q", "[.commits[].oid] | first"))
		summary := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%s", prCommit))
		thisBranchCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log", "--grep", "^"+regexp.QuoteMeta(summary)+"$", "--format=%h"))
		if thisBranchCommit == "" {
			panic(fmt.Sprint("Could not find associated commit for PR (\"", summary, "\") in "+ex.GetMainBranch()))
		}
		info = BranchInfo{CommitHash: thisBranchCommit, BranchName: branchName}
		slog.Info("Using pull request " + commitIndicator + ", commit " + info.CommitHash + ", branch " + info.BranchName)
	case IndicatorTypeCommit:
		slog.Debug("Using commitIndicator as a commit hash " + commitIndicator)

		info = BranchInfo{CommitHash: commitIndicator, BranchName: getBranchForCommit(commitIndicator)}
		slog.Info("Using commit " + info.CommitHash + ", branch " + info.BranchName)
	case IndicatorTypeList:
		slog.Debug("Using commitIndicator as a list index " + commitIndicator)
		newCommits := GetNewCommits(ex.GetMainBranch(), GetCurrentBranchName())
		listIndex, err := strconv.Atoi(commitIndicator)
		if err != nil {
			panic("When indicator type is " + string(IndicatorTypeList) + " commit indicator must be a number, given " + commitIndicator)
		}
		// list indicators are 1 based, convert to 0 based.
		listIndex--
		if listIndex >= len(newCommits) || listIndex < 0 {
			panic("list index " + fmt.Sprint(listIndex) +
				" (parsed from " + commitIndicator + ") " +
				"out of bounds for list of new commits with size " +
				fmt.Sprint(len(newCommits)))
		}
		slog.Info("Using list index " + commitIndicator + ", commit " + newCommits[listIndex].Commit + " " + newCommits[listIndex].Subject)
		info = BranchInfo{CommitHash: newCommits[listIndex].Commit, BranchName: newCommits[listIndex].Branch}
	default:
		panic("Impossible: guessIndicatorType only returns known values, " + fmt.Sprint(indicatorType))
	}
	return info
}

func guessIndicatorType(commitIndicator string) IndicatorType {
	if _, err := strconv.Atoi(commitIndicator); err == nil {
		if len(commitIndicator) < 3 {
			return IndicatorTypeList
		}
		if len(commitIndicator) < 7 {
			return IndicatorTypePr
		}
	}
	return IndicatorTypeCommit
}

func getBranchForCommit(commit string) string {
	sanitizedSubject := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%f", commit))
	return GetBranchForSantizedSubject(sanitizedSubject)
}

func GetBranchForSantizedSubject(sanitizedSubject string) string {
	name := runTemplate("branch-name.template", branchNameTemplateText, getBranchTemplateData(sanitizedSubject))
	if ex.GetMainBranch() == "master" {
		name = strings.Replace(name, "/", "-", -1)
		name = strings.Replace(name, ".", "-", -1)
	}
	name = truncateString(name, 120)
	return name
}

func truncateString(str string, maxBytes int) string {
	for i := range str {
		if i >= maxBytes {
			return str[:i]
		}
	}
	return str
}

func GetPullRequestText(commitHash string, featureFlag string) PullRequestText {
	data := getPullRequestTemplateData(commitHash, featureFlag)
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
			panic(fmt.Sprint("Could not parse ", *configFile, err))
		}
	} else {
		parsed, err = template.New("").Parse(defaultTemplateText)
		if err != nil {
			panic(fmt.Sprint("Could not parse ", defaultTemplateText, err))
		}
	}
	var output bytes.Buffer
	parsed.Execute(&output, data)
	return output.String()
}

func getPullRequestTemplateData(commitHash string, featureFlag string) templateData {
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

func getBranchTemplateData(sanitizedSummary string) branchTemplateData {
	return branchTemplateData{
		Username:             GetUsername(),
		CommitSummaryCleaned: sanitizedSummary,
	}
}

func GetUsername() string {
	if userEmail == "" {
		userEmailRaw := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "user.email"))
		userEmail = userEmailRaw[0:strings.Index(userEmailRaw, "@")]
	}
	return userEmail
}

func getConfigFile(filenameWithoutPath string) *string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprint("Could not get home dir", err))
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
	Branch  string
}

func newGitLogs(logsRaw string) []GitLog {
	logLines := strings.Split(strings.TrimSpace(logsRaw), "\n")
	var logs []GitLog
	for _, logLine := range logLines {
		components := strings.Split(logLine, formatDelimiter)
		if len(components) != 3 {
			// No git logs.
			continue
		}
		logs = append(logs, GitLog{Commit: components[0], Subject: components[1], Branch: GetBranchForSantizedSubject(components[2])})
	}
	return logs
}

func GetAllCommits() []GitLog {
	gitArgs := []string{"--no-pager", "log", "--pretty=format:%h" + formatDelimiter + "%s" + formatDelimiter + "%f", "--abbrev-commit"}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

func GetNewCommits(compareFromRemoteBranch string, to string) []GitLog {
	gitArgs := []string{"--no-pager", "log", "--pretty=format:%h" + formatDelimiter + "%s" + formatDelimiter + "%f", "--abbrev-commit"}
	if RemoteHasBranch(compareFromRemoteBranch) {
		gitArgs = append(gitArgs, "origin/"+compareFromRemoteBranch+".."+to)
	} else {
		gitArgs = append(gitArgs, to)
	}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	return newGitLogs(logsRaw)
}

// Returns most recent commit of the given branch that is on origin/main, or "" if the main branch is not on remote.
func FirstOriginMainCommit(branchName string) string {
	if !LocalHasBranch(branchName) {
		panic("Branch does not exist " + branchName)
	}
	// Verify that remote has branch, there is no origin commit.
	if !RemoteHasBranch(ex.GetMainBranch()) {
		return ""
	}
	return strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "merge-base", "origin/"+ex.GetMainBranch(), branchName))
}

func RemoteHasBranch(branchName string) bool {
	remoteBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-r", "--list", "origin/"+branchName))
	return remoteBranch != ""
}

func LocalHasBranch(branchName string) bool {
	localBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "--list", branchName))
	return localBranch != ""
}

func RequireMainBranch() {
	if GetCurrentBranchName() != ex.GetMainBranch() {
		panic("Must be run from " + ex.GetMainBranch() + " branch")
	}
}

func RequireCommitOnMain(commit string) {
	if commit == ex.GetMainBranch() {
		return
	}
	newCommits := GetNewCommits(ex.GetMainBranch(), "HEAD")
	if !slices.ContainsFunc(newCommits, func(gitLog GitLog) bool {
		return gitLog.Commit == commit
	}) {
		panic("Commit " + commit + " does not exist on " + ex.GetMainBranch() + ". Check `sd log` for available commits.")
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
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "pop")
		log.Println("Popped stash back")
	}
}
