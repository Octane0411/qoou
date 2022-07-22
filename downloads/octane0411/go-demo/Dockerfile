# syntax=docker/dockerfile:1

FROM golang:1.18

WORKDIR $GOPATH/src/github.com/octane0411/go-demo
COPY . $GOPATH/src/github.com/octane0411/go-demo
RUN go build .
EXPOSE 8080
ENTRYPOINT ["./go-demo"]