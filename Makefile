SHELL       := bash
.SHELLFLAGS := -e -o pipefail -c
MAKEFLAGS   += --warn-undefined-variables

export GORACE="halt_on_error=1"

all: version generate lint test

build-cli:
	go build ./cmd/ugo

.PHONY: test
test: version generate lint
	go test -count=1 -cover ./...
	go test -count=1 -race -coverpkg=./... ./...
	go run cmd/ugo/main.go -timeout 20s cmd/ugo/testdata/fibtc.ugo
	go run -race cmd/ugo/main.go -timeout 20s cmd/ugo/testdata/fibtc.ugo

.PHONY: test-long
test-long: version generate lint
	UGO_LONG_TESTS=1 go test -count=1 -cover ./...
	UGO_LONG_TESTS=1 go test -count=1 -race -coverpkg=./... ./...
	go run cmd/ugo/main.go -timeout 20s cmd/ugo/testdata/fibtc.ugo
	go run -race cmd/ugo/main.go -timeout 20s cmd/ugo/testdata/fibtc.ugo

.PHONY: generate-all
generate-all: generate generate-docs

.PHONY: generate
generate: version
	go generate ./...

.PHONY: lint
lint: version
	go vet ./...

.PHONY: generate-docs
generate-docs: version
	go run ./cmd/ugodoc ./stdlib/time ./docs/stdlib-time.md
	go run ./cmd/ugodoc ./stdlib/fmt ./docs/stdlib-fmt.md
	go run ./cmd/ugodoc ./stdlib/strings ./docs/stdlib-strings.md
	go run ./cmd/ugodoc ./stdlib/json ./docs/stdlib-json.md

.PHONY: version
version:
	@go version

.PHONY: clean
clean:
	find . -type f \( -name "cpu.out" -o -name "*.test" -o -name "mem.out" \) -delete
	rm -f cmd/ugo/ugo cmd/ugo/ugo.exe

