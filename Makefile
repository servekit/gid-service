.PHONY: all build run test lint fmt vet proto tidy

## build: Build the server binary
build:
	go build -o bin/server ./cmd/server/

## run: Run the server locally
run:
	go run ./cmd/server/

## test: Run tests with race detector
test:
	go test -race -coverprofile=coverage.out ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	gofmt -w .
	goimports -w .

## vet: Run go vet
vet:
	go vet ./...

## proto: Generate protobuf code with buf
proto:
	buf generate

## tidy: Run go mod tidy
tidy:
	go mod tidy

## all: Format, vet, lint, test
all: fmt vet lint test
