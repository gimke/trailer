BINARY = trailer
GOARCH = amd64

VERSION = v1.0.0
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Symlink into GOPATH
GITHUB_USERNAME=gimke
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

# Build the project
all: clean linux darwin

linux:
	cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ./build/linux-${VERSION}/${BINARY} . ; \
	cd - >/dev/null

darwin:
	cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ./build/darwin-${VERSION}/${BINARY} . ; \
	cd - >/dev/null

#windows:
#	cd ${BUILD_DIR}; \
#	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ./build/windows/${BINARY}.exe . ; \
#	cd - >/dev/null

fmt:
	cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

clean:
	-rm -rf ./build/linux/*
	-rm -rf ./build/darwin/*

.PHONY: link linux darwin windows test vet fmt clean