BINARY := mlbme
VERSION := $(shell cat ./VERSION)
LDFLG := -ldflags "-X main.version=$(VERSION)"

all: build

build:
	go build ${LDFLG} -mod=vendor -v

install:
	go install ${LDFLG} -mod=vendor -v

image: 
	docker build . -t dtpoole/${BINARY} --build-arg=VERSION=${VERSION}

clean:
	rm $(BINARY)

fmt:
	go fmt

.PHONY: build clean install fmt image