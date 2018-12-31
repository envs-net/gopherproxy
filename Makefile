.PHONY: dev build profile bench test clean

all: dev

dev: build
	./gopherproxy -bind 127.0.0.1:8000

build: clean
	go build -o ./gopherproxy ./cmd/gopherproxy/main.go

profile:
	@go test -cpuprofile cpu.prof -memprofile mem.prof -v -bench .

bench:
	@go test -v -bench .

test:
	@go test -v -race -cover -coverprofile=coverage.txt -covermode=atomic .

clean:
	@git clean -f -d -X
