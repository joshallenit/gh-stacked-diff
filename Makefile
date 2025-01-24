.PHONY: build

# TODO: change PATH to include bin for test

format:
	cd src/go; gofmt -w .

build: format
	mkdir -p bin; cd src/go; go build -o ../../bin ./...  

test: build
	cd src/go; go test ./...

release: build
ifndef RELEASE_VERSION
	$(error RELEASE_VERSION is not set)
endif
	rm -rf build/zip
	mkdir -p build/zip/stacked-diff-workflow-bin
	cp bin/* build/zip/stacked-diff-workflow-bin
	cd build/zip; \
	zip -vr stacked-diff-workflow-bin-$(RELEASE_VERSION).zip stacked-diff-workflow-bin/ -x "*.DS_Store"
