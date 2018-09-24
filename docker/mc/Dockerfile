FROM golang:1.11
ADD . /go/src/github.com/himetani/metrics-collector
WORKDIR /go/src/github.com/himetani/metrics-collector
ENV GO111MODULE=on
RUN go get -v -d ./... && \
	make build-mc

FROM ubuntu:18.04
WORKDIR /root/
COPY --from=0 /go/src/github.com/himetani/metrics-collector/cmd/mc/mc .
CMD ["./mc"]
