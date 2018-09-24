.PHONY: all build build-linux install clean test

all: ;

NAME := metrics-collector
REPOHOME := github.com/himetani/metrics-collector
VERSION  := 0.0.9
REVISION  := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags="-s -w"

SRCS    := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

cmd/mc/mc: $(SRCS)
	@echo "=> cd cmd/mc;GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) ./..."
	@cd cmd/mc; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) ./...

cmd/mc/mcweb: $(SRCS)
	@echo "=> cd cmd/mc;GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) ./..."
	@cd cmd/mcweb; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) ./...

$$GOPATH/bin/mc:
	go install $(LDFLAGS)

$$GOPATH/bin/mcweb:
	go install $(LDFLAGS)

build-mc: cmd/mc/mc

build-mcweb: cmd/mc/mcweb

install: $$GOPATH/bin/mc $$GOPATH/bin/mcweb

clean:
	rm -rf bin/*

test: 
	go test -cover -v $(REPOHOME)/...

run:
	@echo "=> find . -type f -name '*go' | grep -v test | xargs go run"
	@find . -type f -name '*go' | grep -v test | xargs go run

exec:
	@echo "=> ./bin/metrics-collector"
	@./bin/metrics-collector
