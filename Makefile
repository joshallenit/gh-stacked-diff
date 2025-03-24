LATEST_VERSION := $(shell grep "latestVersion" "project.properties" | cut -d '=' -f2)
STABLE_VERSION := $(shell grep "stableVersion" "project.properties" | cut -d '=' -f2)
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

ifeq ($(OS),Windows_NT)
	PATH_SEPARATOR:=;
else
	PATH_SEPARATOR:=:
endif
# Add executable to PATH so it can be used as the git sequence editor for unit tests.
export PATH := ${PATH}${PATH_SEPARATOR}${ROOT_DIR}/bin

.PHONY: format
format:
	gofmt -w .
# Note: Using * instead of + in regex so it works on both Windows and Mac.
	sed -i 's/gh-stacked-diff\\/v2@v2\.[0-9]*\.[0-9]*/gh-stacked-diff\\/v2@v'${STABLE_VERSION}'/' README.md

.PHONY: build
build: format
	mkdir -p bin
	go build -o bin

# Example TEST_ARGS:
# make TEST_ARGS="-timeout 30s -run TestSdUpdate_WhenCherryPickCommitsNotSpecified_CherryPicsUserSelection" test
.PHONY: test
test: build
	go test ${TEST_ARGS} ./...
