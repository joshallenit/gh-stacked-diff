package stacked_diff

import (
	"log"
)

type TestExecutor struct {
	Dir        string
	TestLogger *log.Logger
}

func (t TestExecutor) Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	println("TestExecutor.Execute")
	options.Dir = t.Dir
	return DefaultExecutor{}.Execute(options, programName, args...)
}

func (t TestExecutor) Logger() *log.Logger {
	return t.TestLogger
}
