.PHONY: all dev build test vet fmt clean

all: build

dev: build
	./gopherproxy -bind 127.0.0.1:8000

build:
	go build -trimpath -o ./gopherproxy ./cmd/gopherproxy

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w *.go cmd/gopherproxy/*.go

clean:
	rm -f gopherproxy
