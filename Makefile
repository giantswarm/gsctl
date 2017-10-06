PROJECT=gsctl
ORGANISATION=giantswarm
BIN=$(PROJECT)
GOVERSION := 1.8
BUILDDATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse HEAD | cut -c1-10)
SOURCE=$(shell find . -name '*.go')

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

# binary to test with
TESTBIN := build/bin/${BIN}-${GOOS}-${GOARCH}

.PHONY: clean build test crosscompile

all: build

# build binary for current platform
build: $(SOURCE) build/bin/$(BIN)-$(GOOS)-$(GOARCH)

# install binary for current platform (not expected to work on Win)
install: $(SOURCE) build/bin/$(BIN)-$(GOOS)-$(GOARCH)
	cp build/bin/$(BIN)-$(GOOS)-$(GOARCH) /usr/local/bin/$(BIN)

# build for all platforms
crosscompile: build/bin/$(BIN)-darwin-amd64 build/bin/$(BIN)-linux-amd64 build/bin/$(BIN)-windows-386 build/bin/$(BIN)-windows-amd64

# platform-specific build
build/bin/$(BIN)-darwin-amd64:
	@mkdir -p build/bin
	docker run --rm -v $(shell pwd):/go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		-e GOPATH=/go -e GOOS=darwin -e GOARCH=amd64 -e CGO_ENABLED=0 \
		-w /go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		golang:$(GOVERSION)-alpine go build -a -installsuffix cgo -o build/bin/$(BIN)-darwin-amd64 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# platform-specific build for linux-amd64
# - here we have CGO_ENABLED=1
build/bin/$(BIN)-linux-amd64:
	@mkdir -p build/bin
	docker run --rm -v $(shell pwd):/go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		-e GOPATH=/go -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=1 \
		-w /go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		golang:$(GOVERSION)-jessie go build -a -installsuffix cgo -o build/bin/$(BIN)-linux-amd64 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# platform-specific build
build/bin/$(BIN)-windows-386:
	@mkdir -p build/bin
	docker run --rm -v $(shell pwd):/go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		-e GOPATH=/go -e GOOS=windows -e GOARCH=386 -e CGO_ENABLED=0 \
		-w /go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		golang:$(GOVERSION)-alpine go build -a -installsuffix cgo -o build/bin/$(BIN)-windows-386 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# platform-specific build
build/bin/$(BIN)-windows-amd64:
	@mkdir -p build/bin
	docker run --rm -v $(shell pwd):/go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		-e GOPATH=/go -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 \
		-w /go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		golang:$(GOVERSION)-alpine go build -a -installsuffix cgo -o build/bin/$(BIN)-windows-amd64 \
		-ldflags "-X 'github.com/giantswarm/gsctl/config.Version=$(VERSION)' -X 'github.com/giantswarm/gsctl/config.BuildDate=$(BUILDDATE)' -X 'github.com/giantswarm/gsctl/config.Commit=$(COMMIT)'"

# run unittests
gotest:
	go test -cover $(shell glide novendor)

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

	# @${TESTBIN} ping >> /dev/null && echo "OK"
	@${TESTBIN} info >> /dev/null && echo "OK"

# Create binary files for releases
bin-dist: crosscompile

	@mkdir -p bin-dist

	@# test if code signing certificate is available in ./certs/code-signing.p12
	test -f ./certs/code-signing.p12

	@# test if code signing bundle password is in $CODE_SIGNING_CERT_BUNDLE_PASSWORD
	@ if [ "${CODE_SIGNING_CERT_BUNDLE_PASSWORD}" = "" ]; then \
	  echo 'Environment variable $$CODE_SIGNING_CERT_BUNDLE_PASSWORD not set'; \
	  exit 1; \
  fi

	for OS in darwin-amd64 linux-amd64; do \
		mkdir -p build/$(BIN)-$(VERSION)-$$OS; \
		cp README.md build/$(BIN)-$(VERSION)-$$OS/; \
		cp LICENSE build/$(BIN)-$(VERSION)-$$OS/; \
		cp build/bin/$(BIN)-$$OS build/$(BIN)-$(VERSION)-$$OS/$(BIN); \
		cd build/; \
		tar -cvzf ./$(BIN)-$(VERSION)-$$OS.tar.gz $(BIN)-$(VERSION)-$$OS; \
		mv ./$(BIN)-$(VERSION)-$$OS.tar.gz ../bin-dist/; \
		cd ..; \
	done

	@# little different treatment for windows,
	@# becaue of .exe suffix and code signing
	for OS in windows-386 windows-amd64; do \
		mkdir -p build/$(BIN)-$(VERSION)-$$OS; \
		cp README.md build/$(BIN)-$(VERSION)-$$OS/; \
		cp LICENSE build/$(BIN)-$(VERSION)-$$OS/; \
		docker run --rm -ti \
		  -v $(shell pwd)/certs:/mnt/certs \
		  -v $(shell pwd)/build:/mnt/binaries \
		  quay.io/giantswarm/signcode-util:latest \
		  sign \
		  -pkcs12 /mnt/certs/code-signing.p12 \
		  -n "gsctl" \
		  -i https://github.com/giantswarm/gsctl \
		  -in /mnt/binaries/bin/$(BIN)-$$OS \
		  -out /mnt/binaries/$(BIN)-$(VERSION)-$$OS/$(BIN).exe \
		  -pass $(CODE_SIGNING_CERT_BUNDLE_PASSWORD); \
		cd build; \
		zip $(BIN)-$(VERSION)-$$OS.zip $(BIN)-$(VERSION)-$$OS/*; \
		mv ./$(BIN)-$(VERSION)-$$OS.zip ../bin-dist/; \
		cd .. ; \
	done


# This automates a release (except for a GitHub release)
release: bin-dist
	# file uploads to S3
	aws s3 cp bin-dist s3://downloads.giantswarm.io/gsctl/$(VERSION)/ --recursive --exclude="*" --include="*.tar.gz" --acl=public-read
	aws s3 cp bin-dist s3://downloads.giantswarm.io/gsctl/$(VERSION)/ --recursive --exclude="*" --include="*.zip" --acl=public-read
	aws s3 cp VERSION s3://downloads.giantswarm.io/gsctl/VERSION --acl=public-read

	# homebrew
	./update-homebrew.sh

	# scoop
	./update-scoop.sh

	# Update version number occurrences in README.md
	#perl -pi -e "s:gsctl\/[0-9]+\.[0-9]+\.[0-9]+\/:gsctl\/${VERSION}\/:g" README.md
	#perl -pi -e "s:gsctl\-[0-9]+\.[0-9]+\.[0-9]+\-:gsctl\-${VERSION}\-:g" README.md

	#@echo ""
	#@echo "README.md has changed. Please commit and push this change using these commands:"
	#@echo ""
	#@echo "  git commit -m \"Updated version number in README.md to ${VERSION}\" README.md"
	#@echo "git push origin master"
	#@echo ""

# remove generated stuff
clean:
	rm -rf bin-dist build release
