PROJECT_NAME := "serviceCommunicator"
PKG := "github.com/kaizer666/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/ | grep -v example)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all build clean test coverage coverhtml lint

all: build

lint: ## Lint the files
	@go fmt ./...
	@go get -u github.com/mgechev/revive
	@revive -config ~/linterConf.toml

test: ## Run unittests
	@go test -coverprofile=coverprofile_1 ${PKG_LIST} -run \^\(TestMainWithoutPermissions\)\$
	@go test -coverprofile=coverprofile_2 ${PKG_LIST} -run \^\(TestMainWithoutEnv\)\$
	@go test -coverprofile=coverprofile_3 ${PKG_LIST} -run \^\(TestMainWithoutEnvWithFiles\)\$

race: ## Run data race detector
	@go test -race -coverprofile=coverprofile_race_1 ${PKG_LIST} -run \^\(TestMainWithoutPermissions\)\$
	@go test -race -coverprofile=coverprofile_race_2 ${PKG_LIST} -run \^\(TestMainWithoutEnv\)\$
	@go test -race -coverprofile=coverprofile_race_3 ${PKG_LIST} -run \^\(TestMainWithoutEnvWithFiles\)\$

coverage: ## Generate global code coverage report
	~/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	~/coverage.sh html;

build: ## Build the binary file
	@mkdir -p ./deploy
	@go build -o ./deploy/${PROJECT_NAME}

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'