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

.PHONY: clean .gobuild build test crosscompile assert-tagged-version

all: .gobuild build

get-deps: .gobuild

# create Go directory and fetch dependencies
.gobuild:
	@mkdir -p $(GS_PATH)
	@rm -f $(GS_PATH)/$(PROJECT) && cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	#builder get dep -b branch-name https://github.com/giantswarm/gsclientgen.git $(GS_PATH)/gsclientgen
	go get -v github.com/giantswarm/gsclientgen
	go get -v github.com/bradfitz/slice
	go get -v github.com/fatih/color
	go get -v github.com/go-resty/resty
	go get -v github.com/howeyc/gopass
	go get -v github.com/inconshreveable/mousetrap
	go get -v github.com/ryanuber/columnize
	go get -v github.com/spf13/cobra/cobra
	go get -v gopkg.in/yaml.v2

# build binary for current platform
build: .gobuild/bin/$(BIN)-$(GOOS)-$(GOARCH)

# install binary for current platform (not expected to work on Win)
install: .gobuild/bin/$(BIN)-$(GOOS)-$(GOARCH)
	cp .gobuild/bin/$(BIN)-$(GOOS)-$(GOARCH) /usr/local/bin/$(BIN)

# build for all platforms
crosscompile: .gobuild/bin/$(BIN)-darwin-amd64 .gobuild/bin/$(BIN)-linux-amd64 .gobuild/bin/$(BIN)-windows-386 .gobuild/bin/$(BIN)-windows-amd64

# platform-specific build
.gobuild/bin/$(BIN)-darwin-amd64:
	mkdir -p .gobuild/bin
	docker run --rm -v $(shell pwd):/usr/code -w /usr/code \
		-e GOPATH=/usr/code/.gobuild -e GOOS=darwin -e GOARCH=amd64 -e CGO_ENABLED=0 \
		golang:$(GOVERSION) go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-darwin-amd64 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# platform-specific build
.gobuild/bin/$(BIN)-linux-amd64:
	docker run --rm -v $(shell pwd):/usr/code -w /usr/code \
		-e GOPATH=/usr/code/.gobuild -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 \
		golang:$(GOVERSION) go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-linux-amd64 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# platform-specific build
.gobuild/bin/$(BIN)-windows-386:
	docker run --rm -v $(shell pwd):/usr/code -w /usr/code \
		-e GOPATH=/usr/code/.gobuild -e GOOS=windows -e GOARCH=386 -e CGO_ENABLED=0 \
		golang:$(GOVERSION) go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-windows-386 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# platform-specific build
.gobuild/bin/$(BIN)-windows-amd64:
	docker run --rm -v $(shell pwd):/usr/code -w /usr/code \
		-e GOPATH=/usr/code/.gobuild -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 \
		golang:$(GOVERSION) go build -a -installsuffix cgo -o .gobuild/bin/$(BIN)-windows-amd64 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

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
bin-dist: crosscompile
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
	for OS in windows-386 windows-amd64; do \
		mkdir -p build/$(BIN)-$(VERSION)-$$OS; \
		cp README.md build/$(BIN)-$(VERSION)-$$OS/; \
		cp LICENSE build/$(BIN)-$(VERSION)-$$OS/; \
		cp .gobuild/bin/$(BIN)-$$OS build/$(BIN)-$(VERSION)-$$OS/$(BIN).exe; \
		cd build; \
		zip $(BIN)-$(VERSION)-$$OS.zip $(BIN)-$(VERSION)-$$OS/*; \
		mv ./$(BIN)-$(VERSION)-$$OS.zip ../bin-dist/; \
		cd .. ; \
	done

assert-tagged-version:
ifeq ($(shell cat VERSION|grep git),)
	@echo "You are not on a tagged version."
	@echo "Perform a 'git checkout <release-tag>' first."
	@exit 1
endif

# This automates a release (except for a GitHub release)
release: assert-tagged-version bin-dist
	# file uploads to S3
	aws s3 cp bin-dist s3://downloads.giantswarm.io/gsctl/$(VERSION)/ --recursive --exclude="*" --include="*.tar.gz" --acl=public-read
	aws s3 cp bin-dist s3://downloads.giantswarm.io/gsctl/$(VERSION)/ --recursive --exclude="*" --include="*.zip" --acl=public-read
	aws s3 cp VERSION s3://downloads.giantswarm.io/gsctl/VERSION --acl=public-read

	# homebrew
	./update-homebrew.sh

	# Update version number occurrences in README.md
	perl -pi -e "s:gsctl\/[0-9]+\.[0-9]+\.[0-9]+\/:gsctl\/${VERSION}\/:g" README.md
	perl -pi -e "s:gsctl\-[0-9]+\.[0-9]+\.[0-9]+\-:gsctl\-${VERSION}\-:g" README.md

	@echo ""
	@echo "README.md has changed. Please commit and push this change using these commands:"
	@echo ""
	@echo "  git commit -m \"Updated version number in README.md to ${VERSION}\" README.md"
	@echo "git push origin master"
	@echo ""

# remove generated stuff
clean:
	rm -rf bin-dist .gobuild release
