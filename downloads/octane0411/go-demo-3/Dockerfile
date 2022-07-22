# syntax=docker/dockerfile:1

FROM golang:1.18

WORKDIR $GOPATH/src/github.com/octane0411/go-demo
COPY . $GOPATH/src/github.com/octane0411/go-demo
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build .
EXPOSE 8080
ENTRYPOINT ["./go-demo"]