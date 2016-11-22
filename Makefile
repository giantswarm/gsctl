PROJECT=gsctl
BIN = $(PROJECT)
BUILD_PATH := $(shell pwd)/.gobuild
GOPATH := $(BUILD_PATH)
GOVERSION := 1.7.3
GS_PATH := "$(BUILD_PATH)/src/github.com/giantswarm"
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse HEAD | cut -c1-10)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

# binary to test with
TESTBIN := .gobuild/bin/${BIN}-${GOOS}-${GOARCH}

.PHONY: clean .gobuild test

all: .gobuild build

get-deps: .gobuild

# create Go directory and fetch dependencies
.gobuild:
	@mkdir -p $(GS_PATH)
	@rm -f $(GS_PATH)/$(PROJECT) && cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	#builder get dep -b branch-name https://github.com/giantswarm/gsclientgen.git $(GS_PATH)/gsclientgen
	go get github.com/giantswarm/gsclientgen
	go get github.com/bradfitz/slice
	go get github.com/fatih/color
	go get github.com/go-resty/resty
	go get github.com/howeyc/gopass
	go get github.com/inconshreveable/mousetrap
	go get github.com/ryanuber/columnize
	go get github.com/spf13/cobra/cobra
	go get gopkg.in/yaml.v2

# build binaries
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
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-darwin-amd64 -ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

	docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=windows \
		-e GOARCH=386 \
		-e CGO_ENABLED=0 \
		-w /usr/code \
		golang:$(GOVERSION) \
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN).exe -ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

	docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-w /usr/code \
		golang:$(GOVERSION) \
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-linux-amd64 -ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# run some tests
test:
	@${TESTBIN} >> /dev/null && echo "OK"
	@${TESTBIN} help >> /dev/null && echo "OK"
	@${TESTBIN} --help >> /dev/null && echo "OK"
	@${TESTBIN} -h >> /dev/null && echo "OK"

	@${TESTBIN} create --help >> /dev/null && echo "OK"
	@${TESTBIN} info --help >> /dev/null && echo "OK"
	@${TESTBIN} list --help >> /dev/null && echo "OK"
	@${TESTBIN} login --help >> /dev/null && echo "OK"
	@${TESTBIN} logout --help >> /dev/null && echo "OK"
	@${TESTBIN} ping --help >> /dev/null && echo "OK"

	@${TESTBIN} ping >> /dev/null && echo "OK"
	@${TESTBIN} info >> /dev/null && echo "OK"

# remove generated stuff
clean:
	rm -rf $(BUILD_PATH)
