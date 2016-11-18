PROJECT=gsctl
BUILD_PATH := $(shell pwd)/.gobuild
GOPATH := $(BUILD_PATH)
GOVERSION := 1.7.3
BIN = $(PROJECT)
GS_PATH := "$(BUILD_PATH)/src/github.com/giantswarm"
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

.PHONY: clean .gobuild

all: .gobuild build

get-deps: .gobuild

.gobuild:
	@mkdir -p $(GS_PATH)
	@rm -f $(GS_PATH)/$(PROJECT) && cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	builder get dep -b first-version https://github.com/giantswarm/go-client-gen.git $(GS_PATH)/go-client-gen
	go get github.com/fatih/color
	go get github.com/go-resty/resty
	go get github.com/howeyc/gopass
	go get github.com/inconshreveable/mousetrap
	go get github.com/ryanuber/columnize
	go get github.com/spf13/cobra/cobra
	go get gopkg.in/yaml.v2
	go get github.com/bradfitz/slice

build:
	mkdir -p .gobuild/bin
	docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code \
		golang:$(GOVERSION) \
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-darwin-amd64 -ldflags "-X main.version=TODO -X main.buildDate=${BUILDDATE}"

	docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=windows \
		-e GOARCH=386 \
		-e CGO_ENABLED=0 \
		-w /usr/code \
		golang:$(GOVERSION) \
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN).exe -ldflags "-X main.version=TODO -X main.buildDate=${BUILDDATE}"

	docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code \
		golang:$(GOVERSION) \
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-linux-amd64 -ldflags "-X main.version=TODO -X main.buildDate=${BUILDDATE}"

clean:
	rm -rf $(BUILD_PATH)
