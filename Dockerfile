FROM golang:alpine AS build

WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/gopherproxy ./cmd/gopherproxy

FROM alpine:latest

WORKDIR /app
COPY --from=build /out/gopherproxy /usr/local/bin/gopherproxy
COPY robots.txt /app/robots.txt

USER 65532:65532
EXPOSE 8000/tcp
ENTRYPOINT ["/usr/local/bin/gopherproxy"]
