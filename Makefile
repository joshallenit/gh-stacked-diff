.PHONY: build

# TODO: Using * instead of + until we can test on Mac because of -r vs -E flag
# https://stackoverflow.com/questions/2019989/how-to-assign-the-output-of-a-command-to-a-makefile-variable

LATEST_VERSION := $(shell grep "latestVersion" "project.properties" | cut -d '=' -f2)
STABLE_VERSION := $(shell grep "stableVersion" "project.properties" | cut -d '=' -f2)

format:
	gofmt -w .
	sed -i 's/stacked-diff.v2@v2\.[0-9]*\.[0-9]*/stacked-diff/v2@v'${STABLE_VERSION}'/' README.md

build: format
	go build ./...; \

test: build
	export PATH=${PATH}:`pwd`/stacked-diff:`pwd`/stacked-diff.exe;\ 
	go test ./...
