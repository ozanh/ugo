SHELL       := bash
.SHELLFLAGS := -e -o pipefail -c
MAKEFLAGS   += --warn-undefined-variables

all: generate lint test

.PHONY: test
test: generate lint
	go test -count=1 ./...
	go test -count=1 -race -cover ./...
	go run cmd/ugo/main.go -timeout 20s cmd/ugo/testdata/fibtc.ugo

.PHONY: generate
generate:
	go generate ./...

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: generate-docs
generate-docs:
	go run ./cmd/ugodoc ./stdlib/time ./docs/stdlib-time.md
	go run ./cmd/ugodoc ./stdlib/fmt ./docs/stdlib-fmt.md
	go run ./cmd/ugodoc ./stdlib/strings ./docs/stdlib-strings.md
	go run ./cmd/ugodoc ./stdlib/json ./docs/stdlib-json.md

.PHONY: clean
clean:
	find . -type f \( -name "cpu.out" -o -name "*.test" -o -name "mem.out" \) -delete
	rm -f cmd/ugo/ugo cmd/ugo/ugo.exe

