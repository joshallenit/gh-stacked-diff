.PHONY: build

LATEST_VERSION := $(shell grep "latestVersion" "project.properties" | cut -d '=' -f2)
STABLE_VERSION := $(shell grep "stableVersion" "project.properties" | cut -d '=' -f2)
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

ifeq ($(OS),Windows_NT)
	PATH_SEPARATOR:=;
else
	PATH_SEPARATOR:=:
endif
# Add executable to PATH as it is required as the git sequence editor for unit tests.
export PATH := ${PATH}${PATH_SEPARATOR}${ROOT_DIR}

# Note: Using * instead of + in regex so it works on both Windows and Mac. See
# https://stackoverflow.com/questions/2019989/how-to-assign-the-output-of-a-command-to-a-makefile-variable

format:
	gofmt -w .
	sed -i 's/stacked-diff\\/v2@v2\.[0-9]*\.[0-9]*/stacked-diff\\/v2@v'${STABLE_VERSION}'/' README.md

build: format
	rm -f stacked-diff
	rm -f stacked-diff.exe
	go build

test: build
	go test ./...

release: test
ifneq (${LATEST_VERSION}, ${STABLE_VERSION})
$(error stableVersion ${STABLE_VERSION} must be upgraded to latestVersion ${LATEST_VERSION} in project.properties)
endif
	go mod tidy
PORCELAIN := $(shell git status --porcelain)
ifneq (${PORCELAIN}, "")
$(error Changes not committed: ${PORCELAIN})
endif
