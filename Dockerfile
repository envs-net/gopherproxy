FROM golang:alpine

EXPOSE 80/tcp

ENTRYPOINT ["/gopherproxy"]
