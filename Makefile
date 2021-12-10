export GO111MODULE=on
export CGO_ENABLED=1

test: ## run tests, will update go.mod
	go test -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

build: ## build the binary
	go build -o czid -ldflags="-X github.com/chanzuckerberg/czid-cli/pkg/auth0.defaultClientID=${AUTH0_CLIENT_ID} -X github.com/chanzuckerberg/czid-cli/pkg/czid.defaultCZIDBaseURL=${CZID_BASE_URL} -X github.com/chanzuckerberg/czid-cli/pkg.Version=${VERSION} -X github.com/chanzuckerberg/czid-cli/pkg/auth0.defaultAuth0Host=${AUTH0_HOST}" .
.PHONY: build

deps:
	go mod tidy
.PHONY: deps

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod
