
build:
	goreleaser build -f .goreleaser.yml --clean --snapshot

.PHONY: unit-tests
unit-tests: 
	rm -rf /tmp/sdcio/sdc-lite/coverage
	mkdir -p /tmp/sdcio/sdc-lite/coverage
	CGO_ENABLED=1 go test -cover -race ./... -v -covermode atomic -args -test.gocoverdir="/tmp/sdcio/sdc-lite/coverage"
