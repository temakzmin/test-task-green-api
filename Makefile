SHELL := /bin/sh

APP_CONFIG ?= config/config.yaml
GOCACHE ?= $(CURDIR)/.cache/go-build
GO := GOCACHE=$(GOCACHE) go
GO_IMAGE ?= golang:1.25-alpine

.PHONY: help run test coverage fmt lint vet tidy check build-linux-bin build-linux-bin-docker build-linux-bin-auto compose-up compose-prebuilt-up compose-down compose-logs hooks-install clean

help:
	@echo "Available commands:"
	@echo "  make run       - run backend server"
	@echo "  make test      - run all tests"
	@echo "  make coverage  - run tests with coverage"
	@echo "  make fmt       - format Go code (gofumpt + goimports)"
	@echo "  make lint      - run golangci-lint"
	@echo "  make vet       - run go vet"
	@echo "  make tidy      - run go mod tidy"
	@echo "  make check     - fmt + lint + test"
	@echo "  make build-linux-bin - build Linux amd64 backend binary to build/server-linux-amd64"
	@echo "  make build-linux-bin-docker - build Linux amd64 backend binary via golang Docker image"
	@echo "  make hooks-install - install lefthook git hooks"
	@echo "  make compose-up   - start backend+frontend via docker compose (detached)"
	@echo "  make compose-prebuilt-up - start prebuilt compose (expects existing build/server-linux-amd64)"
	@echo "  make compose-down - stop docker compose stack"
	@echo "  make compose-logs - follow docker compose logs"
	@echo "  make clean     - clean local build cache"

run:
	@mkdir -p .cache/go-build
	APP_CONFIG=$(APP_CONFIG) $(GO) run ./cmd/server

test:
	@mkdir -p .cache/go-build
	$(GO) test ./...

coverage:
	@mkdir -p .cache/go-build
	$(GO) test ./... -cover

fmt:
	@mkdir -p .cache/go-build
	@command -v gofumpt >/dev/null 2>&1 || { echo "gofumpt is not installed"; exit 1; }
	@command -v goimports >/dev/null 2>&1 || { echo "goimports is not installed"; exit 1; }
	gofumpt -w $$(find cmd internal -name '*.go')
	goimports -w $$(find cmd internal -name '*.go')

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint is not installed"; exit 1; }
	golangci-lint run

vet:
	@mkdir -p .cache/go-build
	$(GO) vet ./...

tidy:
	@mkdir -p .cache/go-build
	$(GO) mod tidy

check: fmt lint test

hooks-install:
	@command -v lefthook >/dev/null 2>&1 || { echo "lefthook is not installed"; exit 1; }
	lefthook install

build-linux-bin:
	@mkdir -p .cache/go-build build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o build/server-linux-amd64 ./cmd/server

build-linux-bin-docker:
	@mkdir -p .cache/go-build build
	docker run --rm \
		-u $$(id -u):$$(id -g) \
		-v $(CURDIR):/src \
		-w /src \
		-e GOCACHE=/src/.cache/go-build \
		$(GO_IMAGE) \
		sh -c 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/server-linux-amd64 ./cmd/server'

build-linux-bin-auto:
	@if command -v go >/dev/null 2>&1; then \
		$(MAKE) build-linux-bin; \
	else \
		echo "go not found in PATH, using $(GO_IMAGE) to build binary"; \
		$(MAKE) build-linux-bin-docker; \
	fi

compose-up:
	docker compose -f docker/docker-compose.yml up --build -d

compose-prebuilt-up:
	@test -f build/server-linux-amd64 || { \
		echo "missing build/server-linux-amd64"; \
		echo "build binary locally and upload it to server before running this target"; \
		exit 1; \
	}
	docker compose -f docker/docker-compose.yml -f docker/docker-compose.prebuilt.yml up --build -d

compose-down:
	docker compose -f docker/docker-compose.yml down

compose-logs:
	docker compose -f docker/docker-compose.yml logs -f

clean:
	rm -rf .cache/go-build
