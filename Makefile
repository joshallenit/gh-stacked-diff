.PHONY: build

# TODO: change PATH to include bin for test
# TODO: support release on windows (zip not on windows)

format:
	gofmt -w .

build: format
	rm -rf bin; \
	mkdir -p bin; \
	go build -o bin ./...  

test: build
	go test ./...

release: test
ifndef PLATFORM
	$(error PLATFORM is not set)
endif
	rm -rf build/zip
	mkdir -p build/zip/stacked-diff-workflow
	cp bin/* build/zip/stacked-diff-workflow
	export RELEASE_VERSION=`grep "releaseVersion" "project.properties" | cut -d '=' -f2`; \
	cd build/zip; \
	zip -vr stacked-diff-workflow-${PLATFORM}-$(RELEASE_VERSION).zip stacked-diff-workflow/ -x "*.DS_Store"
