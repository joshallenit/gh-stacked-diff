.PHONY: build

# TODO: change PATH to include bin for test
# TODO: support release on windows (zip not on windows)

# TODO: Using * instead of + until we can test on Mac because of -r vs -E flag
# https://stackoverflow.com/questions/2019989/how-to-assign-the-output-of-a-command-to-a-makefile-variable

# rebase-main isn't going to work because the sequence_editor scripts are not on path


LATEST_VERSION := $(shell grep "latestVersion" "project.properties" | cut -d '=' -f2)
STABLE_VERSION := $(shell grep "stableVersion" "project.properties" | cut -d '=' -f2)

format:
	gofmt -w .
	sed -i 's/v2@v2.[0-9]*.[0-9]*/v2@v'${STABLE_VERSION}'/' README.md

build: format
	rm -rf bin; \
	mkdir -p bin; \
	go build -o bin ./...; \
	mv bin/stacked-diff* bin/sd  

test: build
	go test ./...

release: test
ifndef PLATFORM
	$(error PLATFORM is not set)
endif
	rm -rf build/zip
	mkdir -p build/zip/stacked-diff
	cp bin/* build/zip/stacked-diff
	cd build/zip; \
	zip -vr stacked-diff-${PLATFORM}-$(LATEST_VERSION).zip stacked-diff/ -x "*.DS_Store"
