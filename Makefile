BINARY := mlbme
VERSION := $(shell cat ./VERSION)
LDFLG := -ldflags "-X main.version=$(VERSION)"

all: build

build:
	go build ${LDFLG} -v

install:
	go install ${LDFLG} -v

image: 
	docker build . -t dtpoole/${BINARY} --build-arg=VERSION=${VERSION}

clean:
	rm $(BINARY)

fmt:
	go fmt

.PHONY: build clean install fmt image