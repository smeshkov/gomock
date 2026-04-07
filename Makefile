.PHONY: deps clean build

TAG=0.11.0
BINARY=gomock
DIST_DIR=_dist
OS=darwin
ARCH=arm64
VERSION=tip
USER_BIN=${HOME}/bin
DATE=`date +%m-%d-%Y-%H-%M-%S`
MOCK=_dist/mock.json

deps:
	go get -u ./...

clean: 
	rm -rf _dist/*
	
build:
	./_bin/build.sh ${OS} ${VERSION} ${ARCH}

checkfmt: ## Check if the code is formatted
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Go code is not formatted. Run 'make fmt'."; \
		gofmt -d .; \
		exit 1; \
	fi

fmt: ## Format the code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run custom-gcl if installed
	@if command -v custom-gcl > /dev/null; then \
		custom-gcl run; \
	else \
		echo "custom-gcl not installed (go to https://github.com/smeshkov/golangci-lint-plugins), skipping..."; \
	fi

test:
	./_bin/test.sh

run:
	go run cmd/app/main.go -mock ${MOCK} -watch -verbose

tag:
	./_bin/tag.sh ${TAG}

install: build
	chmod +x ${DIST_DIR}/${BINARY}_${OS}_${ARCH}_${VERSION}
	mv ${DIST_DIR}/${BINARY}_${OS}_${ARCH}_${VERSION} ${USER_BIN}/${BINARY}