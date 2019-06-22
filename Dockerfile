FROM golang:1.12-stretch as build

RUN apt-get update && apt-get -uy upgrade \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/bgpat/kubevpn/
ENV GO111MODULE=on

ADD go.mod go.sum /go/src/github.com/bgpat/kubevpn/
RUN go mod download

ADD . .
RUN go build ./cmd/server

FROM debian:stretch-slim

RUN apt-get update && apt-get -uy upgrade \
 && apt-get install -y net-tools \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/github.com/bgpat/kubevpn/server /usr/local/bin/server

EXPOSE 3234
CMD ["/usr/local/bin/server"]
