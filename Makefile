BINARY := mlbme
VERSION := $(shell cat ./VERSION)
LDFLG := -ldflags "-X main.version=$(VERSION)"
PLATFORMS := windows linux darwin
os = $(word 1, $@)

all: build

build:
	go build ${LDFLG} -mod=vendor -v

install:
	go install ${LDFLG} -mod=vendor -v

clean:
	rm -rf ./release
	rm $(BINARY)

fmt:
	go fmt

$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -o release/$(BINARY)-$(VERSION)-$(os)-amd64

release: windows linux darwin

.PHONY: $(PLATFORMS) release build clean install fmt