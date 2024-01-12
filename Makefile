.PHONY: all clean generate fmt
.PHONY: tidy get build test up

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v

GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin

TMPDIR ?= $(CURDIR)/.tmp
TOOLSDIR = $(CURDIR)/internal/build

REVIVE_CONF ?= $(TOOLSDIR)/revive.toml
REVIVE_RUN_ARGS ?= -config $(REVIVE_CONF) -formatter friendly
REVIVE ?= $(GO) run -v github.com/mgechev/revive

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

all: get generate tidy build

clean: ; $(info $(M) cleaning…)
	rm -rf $(TMPDIR)

fmt: ; $(info $(M) reformatting sources…)
	$Q find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

tidy: | fmt ; $(info $(M) tidying up…)
	$Q $(GO) mod tidy
	$Q $(GO) vet ./...
	$Q $(REVIVE) $(REVIVE_RUN_ARGS) ./...

get: ; $(info $(M) downloading dependencies…)
	$Q $(GO) get -v -tags tools ./...

build: ; $(info $(M) building…)
	$Q $(GO) build -v ./...

test: ; $(info $(M) building…)
	$Q $(GO) test ./...

up: ; $(info $(M) updating dependencies…)
	$Q $(GO) get -u -v ./...
	$Q $(GO) mod tidy

generate: ; $(info $(M) generating data…)
	$Q git grep -l '^//go:generate' | sort -uV | xargs -r -n1 $(GO) generate $(GOGENERATE_FLAGS)
