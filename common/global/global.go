package global

var TemplateSet = map[string]bool{
	"golang1.18": true,
}

var TemplateMap = map[string][]string{
	"golang1.18": {"octane0411", "go-template"},
}

var TemplateDockerfileMap = map[string]string{
	"golang1.18": `# syntax=docker/dockerfile:1
FROM golang:1.18
WORKDIR $GOPATH/src/github.com/{{.Username}}/{{.RepoName}}
COPY . $GOPATH/src/github.com/{{.Username}}/{{.RepoName}}
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build .
EXPOSE 8080
ENTRYPOINT ["./{{.RepoName}}"]
`,
}
