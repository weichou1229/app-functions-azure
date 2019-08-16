.PHONY: build test clean docker

GO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=examples/azure-export/azure-export
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

examples/azure-export/azure-export:
	$(GO) build -o $@ ./examples/azure-export

docker:
	docker build \
		-f examples/simple-filter-xml/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(GIT_SHA) \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(VERSION)-dev \
		.

test:
	$(GO) test ./... -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

clean:
	rm -f $(MICROSERVICES)
