PROJECT=gsctl
BIN = $(PROJECT)
BUILD_PATH := $(shell pwd)/.gobuild
RELEASE_PATH := $(shell pwd)/release
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

.PHONY: clean .gobuild build test

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
	rm -rf ./build
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
		go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-windows-386.exe -ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

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

# Create binary files for releases
bin-dist:
	rm -rf build

	mkdir bin-dist

	for OS in darwin-amd64 linux-amd64; do \
		mkdir -p build/$(BIN)-$(VERSION)-$$OS; \
		cp README.md build/$(BIN)-$(VERSION)-$$OS/; \
		cp LICENSE build/$(BIN)-$(VERSION)-$$OS/; \
		cp .gobuild/bin/$(BIN)-$$OS build/$(BIN)-$(VERSION)-$$OS/$(BIN); \
		cd build/; \
		tar -cvzf ./$(BIN)-$(VERSION)-$$OS.tar.gz $(BIN)-$(VERSION)-$$OS; \
		mv ./$(BIN)-$(VERSION)-$$OS.tar.gz ../bin-dist/; \
		cd ..; \
	done

	# little different treatment for windows
	mkdir -p build/$(BIN)-$(VERSION)-windows-386
	cp README.md build/$(BIN)-$(VERSION)-windows-386/
	cp LICENSE build/$(BIN)-$(VERSION)-windows-386/
	cp .gobuild/bin/$(BIN)-windows-386.exe build/$(BIN)-$(VERSION)-windows-386/$(BIN).exe
	cd build && zip $(BIN)-$(VERSION)-windows-386.zip $(BIN)-$(VERSION)-windows-386/*
	mv build/$(BIN)-$(VERSION)-windows-386.zip bin-dist/


# This should, at some point, automate releases.
release: bin-dist
	# file uploads to S3
	aws s3 cp bin-dist s3://downloads.giantswarm.io/gsctl/$(VERSION)/ --recursive --exclude="*" --include="*.tar.gz" --acl=public-read
	aws s3 cp bin-dist s3://downloads.giantswarm.io/gsctl/$(VERSION)/ --recursive --exclude="*" --include="*.zip" --acl=public-read
	aws s3 cp VERSION s3://downloads.giantswarm.io/gsctl/VERSION --acl=public-read

# remove generated stuff
clean:
	rm -rf bin-dist $(BUILD_PATH) $(RELEASE_PATH)
