package commands

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"bytes"
	"strings"

	"context"

	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

// Program name
const programName string = "gh-stacked-diff"

// Calls [parseArguments] for unit tests.
func testParseArguments(commandLineArgs ...string) string {
	if slog.Default().Handler().Enabled(context.Background(), slog.LevelInfo) {
		out := testutil.NewWriteRecorder()
		testParseArgumentsWithOut(out, commandLineArgs...)
		return out.String()
	} else {
		out := new(bytes.Buffer)
		testParseArgumentsWithOut(out, commandLineArgs...)
		return out.String()
	}
}

func testParseArgumentsWithOut(out io.Writer, commandLineArgs ...string) {
	slog.Debug(fmt.Sprint("***Testing parse arguments*** ", strings.Join(commandLineArgs, " ")))
	panicOnExit := func(code int) {
		panic("Panicking instead of exiting with code " + fmt.Sprint(code))
	}
	// Executable must be on PATH for tests to pass so that sequenceEditorPrefix will execute.
	// PATH is set in ../Makefile
	appExecutable := programName

	if !slices.ContainsFunc(commandLineArgs, func(next string) bool {
		return strings.HasPrefix(next, "--log-level")
	}) {
		// Use current log level if it set to something other than Info.
		level := lowestSupportedLogLevel()
		if level != slog.LevelInfo {
			loglevelFlag := "--log-level=" + level.String()
			appExecutable += " " + loglevelFlag
			commandLineArgs = slices.Insert(commandLineArgs, 0, loglevelFlag)
		}
	}

	// Set stdin in unit tests to avoid error with bubbletea:
	// "error creating cancelreader: failed to prepare console input: get console mode: The handle is invalid."
	// To fake user input use interactive.SendToProgram.
	stdin := strings.NewReader("")

	appConfig := util.AppConfig{
		Io:            util.StdIo{Out: out, Err: out, In: stdin},
		AppExecutable: appExecutable,
		Exit:          panicOnExit,
		UserCacheDir:  getTestAppCacheDir(),
	}
	parseArguments(
		appConfig,
		flag.NewFlagSet("sd", flag.ContinueOnError),
		commandLineArgs,
	)
	slog.Debug(fmt.Sprint("***Done running arguments*** ", strings.Join(commandLineArgs, " ")))
}

func lowestSupportedLogLevel() slog.Level {
	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn}
	for _, level := range levels {
		if slog.Default().Handler().Enabled(context.Background(), level) {
			return level
		}
	}
	return slog.LevelError
}

func getTestAppCacheDir() string {
	// okay I need it as a C:\\ in order to use WriteFile/ReadFile
	// but all of the path stuff uses /
	wd, err := os.Getwd()
	if err != nil {
		panic("cannot get wd: " + err.Error())
	}
	parentDir, _ := filepath.Split(wd)
	userCacheDir := filepath.Join(parentDir, "user-cache")
	// nolint:errcheck
	err = os.Mkdir(userCacheDir, os.ModePerm)
	return userCacheDir
}
