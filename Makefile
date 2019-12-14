goimports:
	goimports -w .

lint:
	golangci-lint run -v ./...

build:
	mkdir -p ./bin
	go build -o ./bin/ ./cmd/git-fzf

test:
	go test -v ./...

test-cover:
	go test -cover -v ./...

install:
	go install ./cmd/git-fzf
