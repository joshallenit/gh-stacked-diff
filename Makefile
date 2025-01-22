.PHONY: build

test:
	cd src/go; go test ./...

build:
	cd src/go; gofmt -w .
	mkdir -p bin; cd src/go; go build -o ../../bin ./...  

release: build
ifndef RELEASE_VERSION
	$(error RELEASE_VERSION is not set)
endif
	rm -rf build/zip
	mkdir -p build/zip/stacked-diff-workflow-bin
	cp bin/* build/zip/stacked-diff-workflow-bin
	cd build/zip; \
	zip -vr stacked-diff-workflow-bin-$(RELEASE_VERSION).zip stacked-diff-workflow-bin/ -x "*.DS_Store"
