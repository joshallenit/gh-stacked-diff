package execute

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
)

var Red = "\033[31m"
var Reset = "\033[0m"
var White = "\033[97m"
var Yellow = "\033[33m" // #C0A000 (Buddha Gold)

var mainBranchName string

/*
Options for [ExecuteWithOptions].
*/
type ExecuteOptions struct {
	// String to use for stdin. Useful for "git apply".
	Stdin *string
	// For example "MY_VAR=some_value"
	EnvironmentVariables []string
	Output               *ExecutionOutput
}

type ExecutionOutput struct {
	Stdout io.Writer
	Stderr io.Writer
}

func NewStandardOutput() *ExecutionOutput {
	return &ExecutionOutput{Stdout: os.Stdout, Stderr: os.Stderr}
}

type Executor interface {
	Execute(options ExecuteOptions, programName string, args ...string) (string, error)
	Logger() *slog.Logger
}

var globalExecutor Executor = DefaultExecutor{}

type DefaultExecutor struct{}

func SetGlobalExecutor(executor Executor) {
	globalExecutor = executor
}

func (defaultExecutor DefaultExecutor) Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	cmd := exec.Command(programName, args...)
	if options.EnvironmentVariables != nil {
		cmd.Env = append(os.Environ(), options.EnvironmentVariables...)
	}
	if options.Stdin != nil {
		cmd.Stdin = strings.NewReader(*options.Stdin)
	}
	var out []byte
	var err error
	if options.Output != nil {
		cmd.Stdout = options.Output.Stdout
		cmd.Stderr = options.Output.Stderr
		err = cmd.Run()
	} else {
		out, err = cmd.CombinedOutput()
	}

	stringOut := string(out)
	defaultExecutor.Logger().Debug("Executed " + getLogMessage(programName, args, stringOut, err))
	return stringOut, err
}

func (defaultExecutor DefaultExecutor) Logger() *slog.Logger {
	return slog.Default()
}

func Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	return globalExecutor.Execute(options, programName, args...)
}

func ExecuteOrDie(options ExecuteOptions, programName string, args ...string) string {
	out, err := Execute(options, programName, args...)
	if err != nil {
		debug.PrintStack()
		globalExecutor.Logger().Error(Red + "Failed executing " + Reset + getLogMessage(programName, args, out, err))
		os.Exit(1)
	}
	return out
}

func getLogMessage(programName string, args []string, out string, err error) string {
	logMessage := programName + " " + strings.Join(args, " ")
	stringOut := string(out)
	if strings.TrimSpace(stringOut) != "" {
		logMessage = logMessage + "\n" + stringOut
	}
	if err != nil {
		logMessage = logMessage + "\nError " + err.Error()
		var exerr *exec.ExitError
		if errors.As(err, &exerr) {
			logMessage = logMessage + " (Exit code " + fmt.Sprint(exerr.ExitCode()) + ")"
		}
	}
	return logMessage
}

func GetMainBranch() string {
	if mainBranchName == "" {
		remoteMainBranch, err := Execute(ExecuteOptions{}, "git", "rev-parse", "--abbrev-ref", "origin/HEAD")
		if err == nil {
			remoteMainBranch = strings.TrimSpace(remoteMainBranch)
			mainBranchName = remoteMainBranch[strings.Index(remoteMainBranch, "/")+1:]
		} else {
			// Remote is empty, use config.
			mainBranchName = strings.TrimSpace(ExecuteOrDie(ExecuteOptions{}, "git", "config", "init.defaultBranch"))
		}
	}
	return mainBranchName
}
