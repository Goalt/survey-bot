.PHONY: lint test mocks

lint:
	golangci-lint run

test:
	go test ./... -cover

build:
	GOARCH="arm64" go build -o ./bin/cli ./cmd/cli
	GOOS="darwin" GOARCH="arm64"  go build -o ./bin/cli.darwin ./cmd/cli

mocks:
	mockery --name=TelegramRepo --dir=./internal/service --output=./internal/service/mocks
	mockery --name=DBTransaction --dir=./internal/service --output=./internal/service/mocks
	mockery --name=DBRepo --dir=./internal/service --output=./internal/service/mocks
	mockery --name=ResultsProcessor --dir=./internal/entity --output=./internal/service/mocks
	mockery --name=Logger --dir=./internal/logger --output=./internal/service/mocks