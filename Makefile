
build:
	cd cmd/shortener && go build -o shortener *.go

test:
	shortenertest -test.v -test.run=^TestIteration1$$ -binary-path=cmd/shortener/shortener

run:
	go run cmd/shortener/*.go

.PHONY: build test run