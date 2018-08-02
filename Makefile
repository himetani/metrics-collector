.PHONY: all build build-linux install clean test

all: ;

NAME := metrics-collector
REPOHOME := github.com/himetani/metrics-collector
VERSION  := 0.0.9
REVISION  := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags="-s -w"

SRCS    := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

bin/$(NAME): $(SRCS)
	go build $(LDFLAGS) -o bin/$(NAME)

bin/linux/$(NAME): $(SRCS)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/linux/$(NAME)

$$GOPATH/bin/$(NAME):
	go install $(LDFLAGS)

build: bin/$(NAME)

build-linux: bin/linux/$(NAME)

install: $$GOPATH/bin/$(NAME)

clean:
	rm -rf bin/*

test: 
	go test -cover -v $(REPOHOME)/...

run:
	@echo "=> find . -type f -name '*go' | grep -v test | xargs go run"
	@find . -type f -name '*go' | grep -v test | xargs go run
