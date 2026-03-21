.PHONY: build lint test setup

build:
	go build ./...

lint:
	go vet ./...
	golangci-lint run

test:
	go test ./...

setup:
	git config core.hooksPath .githooks
	@echo "Git hooks enabled."
