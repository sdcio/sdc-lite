# go versions
TARGET_GO_VERSION := $(shell awk '/^go / {print "go"$$2}' go.mod)
GO_FALLBACK := go
# We prefer $TARGET_GO_VERSION if it is not available we go with whatever go we find ($GO_FALLBACK)
GO_BIN := $(shell if [ "$$(which $(TARGET_GO_VERSION))" != "" ]; then echo $$(which $(TARGET_GO_VERSION)); else echo $$(which $(GO_FALLBACK)); fi)

build:
	mkdir -p bin
	CGO_ENABLED=0 ${GO_BIN} build -o bin/config-diff .

.PHONY: unit-tests
unit-tests: 
	rm -rf /tmp/sdcio/config-diff/coverage
	mkdir -p /tmp/sdcio/config-diff/coverage
	CGO_ENABLED=1 go test -cover -race ./... -v -covermode atomic -args -test.gocoverdir="/tmp/sdcio/config-diff/coverage"
