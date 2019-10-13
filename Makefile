BINARY := SopsSecretGenerator
VERSION ?= $(shell (git describe --tags --always --dirty --match=v* 2>/dev/null || echo v0) | cut -c2-)
PLATFORMS := windows linux darwin
ARCH := amd64
RELEASE_DIR := release

releases = $(patsubst %,$(RELEASE_DIR)/$(BINARY)_$(VERSION)_%_$(ARCH), $(PLATFORMS))
platform = $(patsubst $(RELEASE_DIR)/$(BINARY)_$(VERSION)_%_$(ARCH),%, $@)


export GO111MODULE=on

$(BINARY): SopsSecretGenerator.go
	go build -o $@ $<

.PHONY: test
test:
	go test -v -race

.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: release
release: $(releases)

$(releases): SopsSecretGenerator.go
	GOOS=$(platform) GOARCH=$(ARCH) go build -o $@ $<
	GOOS=$(platform) GOARCH=$(ARCH) go build -ldflags '-extldflags "-fno-PIC -static"' -tags 'osusergo netgo static_build' -o $@ $<

.PHONY: clean
clean:
	-rm -rf $(BINARY) $(RELEASE_DIR)
