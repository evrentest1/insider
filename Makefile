GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.DEFAULT_GOAL := help
.PHONY: init install-tools generate-mocks delete-mocks docs build help test lint sqlc-generate

all: help

init: install-tools generate-mocks docs ### Prepares repository to work

install-tools: ### Install tools
	@cat cmd/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

generate-mocks: delete-mocks ### Generate mock files after deleting mocks
	go generate ./...

delete-mocks:
	find . -name "*mocks*.go" -delete

docs: ### Build swagger documentation
	swag fmt; swag init --dir internal --parseInternal --parseDependency --parseDepth 4 -g /api/handlers/handler.go

build:
	go build -a -o ./main cmd/main.go

test:
	go test ./...

sqlc-generate: ### Generate Go code from SQL queries
	$(GOPATH)/bin/sqlc generate

lint: ### Run golintci-lint on your project
	golangci-lint run --config .golangci.yml ./...

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)