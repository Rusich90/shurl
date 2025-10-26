
build:
	cd cmd/shortener && go build -o shortener *.go

test:
	SERVER_PORT=8888 shortenertest -test.v -test.run=^TestIteration4$$ -binary-path=cmd/shortener/shortener -server-port=$SERVER_PORT

my-test:
	go test -v ./...

run:
	go run cmd/shortener/*.go

.PHONY: build test run