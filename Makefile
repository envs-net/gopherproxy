.PHONY: dev build clean

all: dev

dev: build
	./gopherproxy -bind 127.0.0.1:8000

build: clean
	go get ./...
	go build -o ./gopherproxy ./cmd/gopherproxy/main.go

clean:
	rm -rf gopherproxy
