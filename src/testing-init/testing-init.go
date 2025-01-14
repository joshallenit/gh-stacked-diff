package testing_init

import (
	"path"
	"runtime"
)

var RootProjectDir string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	RootProjectDir = path.Join(path.Dir(filename), "/../..")
}
