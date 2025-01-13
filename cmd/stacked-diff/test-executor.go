package stacked_diff

type TestExecutor struct {
}

func (t TestExecutor) Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	println("TestExecutor.Execute")
	return DefaultExecutor{}.Execute(options, programName, args...)
}
