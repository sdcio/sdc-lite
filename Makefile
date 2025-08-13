
build:
	goreleaser build -f .goreleaser.yml --clean --snapshot

.PHONY: unit-tests
unit-tests: 
	rm -rf /tmp/sdcio/config-diff/coverage
	mkdir -p /tmp/sdcio/config-diff/coverage
	CGO_ENABLED=1 go test -cover -race ./... -v -covermode atomic -args -test.gocoverdir="/tmp/sdcio/config-diff/coverage"
