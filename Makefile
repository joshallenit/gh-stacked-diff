build:
	go build -o bin/add-reviewers cmd/add-reviewers.go cmd/execute.go cmd/templates.go
	go build -o bin/new-pr cmd/new-pr.go cmd/execute.go cmd/templates.go
	go build -o bin/replace-commit cmd/replace-commit.go cmd/execute.go cmd/templates.go
	go build -o bin/replace-head cmd/replace-head.go cmd/execute.go cmd/templates.go
	go build -o bin/gitlog cmd/gitlog.go cmd/execute.go cmd/templates.go
