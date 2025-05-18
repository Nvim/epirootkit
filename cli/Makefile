.PHONY: build
build:
	go build -o=./bin/cli ./cmd/main.go

.PHONY: run
run: build
	./bin/cli

.PHONY: tidy
tidy:
	go mod tidy -v

format:
	gofumpt -w ./**/*.go

lint:
	golangci-lint run
