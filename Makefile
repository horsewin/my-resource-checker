.PHONY: build test clean run install lint fmt

BINARY_NAME=sbcntr-validator
MAIN_PATH=main.go

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

test:
	go test -v ./...

test-coverage:
	go test -v -cover ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

run:
	go run $(MAIN_PATH)

install:
	go mod download

lint:
	golangci-lint run

fmt:
	go fmt ./...

build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)