GOLANGCI_LINT_VERSION ?= v2.12.2

install:
	go install -v ./cmd/monero

build:
	go build -v ./cmd/monero

test:
	go test -v ./pkg/...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --config=.golangci.yaml

image:
	docker build -t go-monero .

.PHONY: install build test lint image
