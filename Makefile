export GO111MODULE=on

SopsSecretGenerator: SopsSecretGenerator.go
	go build -o $@ $<

.PHONY: test
test:
	go test -v -race

.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: release
release:
	goreleaser --rm-dist --skip-publish

.PHONY: publish-release
publish-release:
	goreleaser --rm-dist

.PHONY: clean
clean:
	-rm -rf $(BINARY) $(RELEASE_DIR)
