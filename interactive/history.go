package interactive

import (
	"os"
	"path"
	"slices"
	"strings"
)

const MAX_HISTORY = 30

/*
history items are returned as:
[0] least recent
[last element] most recent
*/
func readHistory() []string {
	data, err := os.ReadFile(getHistoryFile())
	if err != nil {
		return []string{}
	}
	return strings.Split(string(data), "\n")
}

// Add a most recently used item to history.
func addToHistory(history []string, newHistoryItem string) {
	// remove any duplicates
	history = slices.DeleteFunc(history, func(next string) bool {
		return next == newHistoryItem
	})
	history = append(history, newHistoryItem)
	if len(history) > MAX_HISTORY {
		history = history[len(history)-MAX_HISTORY:]
	}
	data := strings.Join(history, "\n")
	if writeErr := os.WriteFile(getHistoryFile(), []byte(data), os.ModePerm); writeErr != nil {
		panic(writeErr)
	}
}

func getHistoryFile() string {
	return path.Join(getAppCacheDir(), "reviewers_history.txt")
}

func getAppCacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic("Cannot find UserCacheDir: " + err.Error())
	}
	appCacheDir := path.Join(userCacheDir, "gh-stacked-diff")
	// nolint:errcheck
	os.Mkdir(appCacheDir, os.ModePerm)
	return appCacheDir
}

func allUsersFromHistory(history []string) []string {
	allUsers := make([]string, 0, len(history))
	for _, next := range history {
		users := strings.FieldsFunc(next, func(next rune) bool {
			return slices.Contains(getBreakingChars(), next)
		})
		allUsers = slices.AppendSeq(allUsers, slices.Values(users))
	}
	slices.Sort(allUsers)
	return slices.Compact(allUsers)
}

func getBreakingChars() []rune {
	return []rune{' ', ','}
}
