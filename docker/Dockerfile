FROM golang:1.12-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
RUN git clone https://github.com/ts2/ts2-sim-server
RUN go get github.com/GeertJohan/go.rice/rice
ENV GO111MODULE=on
RUN cd ts2-sim-server && GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/ts2-sim-server
RUN rice append --exec /go/bin/ts2-sim-server -i ./ts2-sim-server/server

FROM alpine:latest
MAINTAINER Nicolas Piganeau <npi@m4x.org>
COPY demo.json /
COPY --from=builder /go/bin/ts2-sim-server /ts2-sim-server
CMD ["/ts2-sim-server", "-loglevel", "debug", "/demo.json"]
