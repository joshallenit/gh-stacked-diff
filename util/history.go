package util

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const MAX_HISTORY = 30

/*
history items are returned as:
[0] least recent
[last element] most recent
*/
func ReadHistory(appConfig AppConfig, historyFilename string) []string {
	data, err := os.ReadFile(getHistoryFile(appConfig, historyFilename))
	if err != nil {
		return []string{}
	}
	return strings.Split(string(data), "\n")
}

// Add a most recently used item to history.
func AddToHistory(history []string, newHistoryItem string) []string {
	// remove any duplicates
	history = slices.DeleteFunc(history, func(next string) bool {
		return next == newHistoryItem
	})
	return append(history, newHistoryItem)
}

// Add a most recently used item to history.
func SetHistory(appConfig AppConfig, historyFileName string, history []string) {
	if len(history) > MAX_HISTORY {
		history = history[len(history)-MAX_HISTORY:]
	}
	data := strings.Join(history, "\n")
	if writeErr := os.WriteFile(getHistoryFile(appConfig, historyFileName), []byte(data), os.ModePerm); writeErr != nil {
		panic("Could not write file: " + writeErr.Error())
	}
}

func getHistoryFile(appConfig AppConfig, historyFilename string) string {
	appCacheDir := filepath.Join(appConfig.UserCacheDir, "gh-stacked-diff", GetRepoName())
	ExecuteOrDie(ExecuteOptions{}, "mkdir", "-p", appCacheDir)
	return filepath.Join(appCacheDir, historyFilename)
}
