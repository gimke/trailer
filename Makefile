BINARY = trailer
GOARCH = amd64

VERSION = $(shell cat .ver)
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Symlink into GOPATH
GITHUB_USERNAME=gimke
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION}"

# Build the project
all: clean linux darwin

linux:
	@GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ./build/linux-${VERSION}/${BINARY} .
	@cd ${BUILD_DIR} && /bin/echo -n "${VERSION}" > ./build/linux-${VERSION}/.ver
	@cd ${BUILD_DIR}/build && zip -q -r linux-${VERSION}.zip ./linux-${VERSION}/*
	@echo "\033[32;1mBuild linux Done \033[0m"

darwin:
	@GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ./build/darwin-${VERSION}/${BINARY} .
	@cd ${BUILD_DIR} && /bin/echo -n "${VERSION}" > ./build/darwin-${VERSION}/.ver
	@cd ${BUILD_DIR}/build && zip -q -r darwin-${VERSION}.zip ./darwin-${VERSION}/*
	@echo "\033[32;1mBuild darwin Done \033[0m"

#windows:
#	cd ${BUILD_DIR}; \
#	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ./build/windows/${BINARY}.exe . ; \
#	cd - >/dev/null

fmt:
	@cd ${BUILD_DIR}
	@go fmt $$(go list ./... | grep -v /vendor/)

clean:
	@rm -rf ./build/linux-${VERSION}.zip
	@rm -rf ./build/darwin-${VERSION}.zip
	@rm -rf ./build/linux-${VERSION}/*
	@rm -rf ./build/darwin-${VERSION}/*