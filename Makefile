BINARY := SopsSecret
VERSION ?= $(shell (git describe --tags --always --dirty --match=v* 2>/dev/null || echo v0) | cut -c2-)
PLATFORMS := windows linux darwin
ARCH := amd64
RELEASE_DIR := release

releases = $(patsubst %,$(RELEASE_DIR)/$(BINARY)_$(VERSION)_%_$(ARCH), $(PLATFORMS))
platform = $(patsubst $(RELEASE_DIR)/$(BINARY)_$(VERSION)_%_$(ARCH),%, $@)


export GO111MODULE=on

$(BINARY): SopsSecret.go
	go build -o $@ $<

.PHONY: test
test:
	go test -v -race

.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: release
release: $(releases)

$(releases): SopsSecret.go
	GOOS=$(platform) GOARCH=$(ARCH) go build -o $@ $<

.PHONY: clean
clean:
	-rm -rf $(BINARY) $(RELEASE_DIR)
