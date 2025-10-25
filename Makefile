
build:
	cd cmd/shortener && go build -o shortener *.go

test:
	shortenertest -test.v -test.run=^TestIteration3$$ -source-path=.

my-test:
	go test -v ./...

run:
	go run cmd/shortener/*.go

.PHONY: build test run