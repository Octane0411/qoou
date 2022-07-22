# syntax=docker/dockerfile:1
FROM golang:1.18
WORKDIR $GOPATH/src/github.com/YanQiaoQi/test
COPY . $GOPATH/src/github.com/YanQiaoQi/test
RUN go build .
EXPOSE 8080
ENTRYPOINT ["./test"]
