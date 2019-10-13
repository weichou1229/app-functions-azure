.PHONY: build test clean docker

GO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=examples/azure-export/azure-export examples/http-command-service/http-command
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

examples/azure-export/azure-export:
	$(GO) build -o $@ ./examples/azure-export

examples/http-command-service/http-command:
	$(GO) build -o $@ ./examples/http-command-service

test:
	$(GO) test ./... -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

clean:
	rm -f $(MICROSERVICES)
