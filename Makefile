.PHONY: vhs-rec vhs-out ci

# Replace this by nix or similar
install:
	go install mvdan.cc/gofumpt@v0.8.0
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0

ci:
	golangci-lint run ./...
	gofumpt -l -w .
	go vet ./...
	go test -race -cover ./...
	go mod tidy
	go build ./...
