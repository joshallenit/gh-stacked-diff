.PHONY: build

build:
	gofmt -w cmd
	go build -o bin/add-reviewers cmd/add-reviewers.go cmd/execute.go cmd/templates.go
	go build -o bin/get-branch-name-for cmd/get-branch-name-for.go cmd/execute.go cmd/templates.go
	go build -o bin/get-main-branch cmd/get-main-branch.go cmd/execute.go cmd/templates.go
	go build -o bin/gitlog cmd/gitlog.go cmd/execute.go cmd/templates.go
	go build -o bin/new-pr cmd/new-pr.go cmd/execute.go cmd/templates.go
	go build -o bin/replace-commit cmd/replace-commit.go cmd/execute.go cmd/templates.go
	go build -o bin/replace-head cmd/replace-head.go cmd/execute.go cmd/templates.go
	go build -o bin/sequence-editor-mark-as-fixup cmd/sequence-editor-mark-as-fixup.go cmd/execute.go cmd/templates.go
	go build -o bin/update-pr cmd/update-pr.go cmd/execute.go cmd/templates.go
	go build -o bin/wait-for-merge cmd/wait-for-merge.go cmd/execute.go cmd/templates.go

release: build
ifndef RELEASE_VERSION
	$(error RELEASE_VERSION is not set)
endif
	rm -rf build/zip
	mkdir -p build/zip/stacked-diff-workflow-bin
	cp bin/* build/zip/stacked-diff-workflow-bin
	cd build/zip; \
	zip -vr stacked-diff-workflow-bin-$(RELEASE_VERSION).zip stacked-diff-workflow-bin/ -x "*.DS_Store"
