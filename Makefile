PROJECT_NAME := "bitwarden-alfred-workflow"
PKG := "github.com/blacs30/$(PROJECT_NAME)"
GO111MODULE=on
.EXPORT_ALL_VARIABLES:
.PHONY: all dep lint vet test test-coverage build clean

all: build copy-build-assets

dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golangci-lint run --timeout 3m

vet: ## Run go vet
	@go vet ./src

test: ## Run unittests
	@go test -short ./src

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ./src
	@cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o workflow/$(PROJECT_NAME)-amd64 ./src
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o workflow/$(PROJECT_NAME) ./src

universal-binary:
	@lipo -create -output workflow/bitwarden-alfred-workflow workflow/bitwarden-alfred-workflow-amd64 workflow/bitwarden-alfred-workflow-arm64

clean: ## Remove previous build
	@rm -f workflow/$(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-hooks:
	@mkdir -p .git/hooks
	@cp .github/hooks/* .git/hooks
	@chmod +x .git/hooks/*

copy-build-assets:
	@chmod +x ./workflow/*.sh
	@cp -r assets ./workflow
	@go install github.com/pschlump/markdown-cli
	@markdown-cli -i README.md -o workflow/README.html

package-alfred:
	@cd ./workflow && zip -r ../bitwarden-alfred-workflow.alfredworkflow ./* \
	&& cd .. && rm -rf workflow && git checkout workflow
