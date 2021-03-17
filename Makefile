export GO111MODULE=on
export CGO_ENABLED=1

test: ## run tests, will update go.mod
	go test -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

build: ## build the binary
	go build -o idseq -ldflags="-X github.com/chanzuckerberg/idseq-cli-v2/pkg/auth0.clientID=${AUTH0_CLIENT_ID} -X github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq.roleARN=${ROLE_ARN} -X github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq.baseURL=${IDSEQ_BASE_URL}" .
.PHONY: build

deps:
	go mod tidy
.PHONY: deps

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod
