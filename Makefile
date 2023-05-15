.PHONY: deps clean build

TAG=0.9.3
BINARY=gomock
DIST_DIR=_dist
OS=darwin
ARCH=arm64
VERSION=tip
USER_BIN=${HOME}/bin
DATE=`date +%m-%d-%Y-%H-%M-%S`
MOCK=mock.json

deps:
	go get -u ./...

clean: 
	rm -rf _dist/*
	
build:
	./_bin/build.sh ${OS} ${VERSION} ${ARCH}

test:
	./_bin/test.sh

# run:
# 	./_dist/${BINARY}_${OS}_${ARCH}_${VERSION} -mock ${MOCK}

run:
	go run cmd/app/main.go -mock ${MOCK} -verbose

tag:
	./_bin/tag.sh ${TAG}

# install:
# 	./_bin/install.sh ${OS} ${ARCH}

install: build
	chmod +x ${DIST_DIR}/${BINARY}_${OS}_${ARCH}_${VERSION}
	mv ${DIST_DIR}/${BINARY}_${OS}_${ARCH}_${VERSION} ${USER_BIN}/${BINARY}