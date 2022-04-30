package v1

import (
	"bytes"
	"fmt"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/util"
	"net/http"
	"time"
)

var Cli = &http.Client{}

type GitHubCommitResponse []GitHubCommit

type TokenRequest struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type GitHubCreateRepoRequest struct {
	Name              string `json:"name"`
	AutoInit          bool   `json:"auto_init"`
	Private           bool   `json:"private"`
	GitignoreTemplate string `json:"gitignore_template"`
}

type CreateRepoRequest struct {
	Username string `json:"username"`
	RepoName string `json:"repoName"`
}

type CreateRepoWithTemplateRequest struct {
	Username string `json:"username"`
	RepoName string `json:"repoName"`
	Template string `json:"template"`
}

type GitHubErrorResponse struct {
	Message string `json:"message"`
}

type DeployRequest struct {
	Username string `json:"username"`
	RepoName string `json:"repoName"`
}

type GitHubCommit struct {
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

func GetGitHubCommitsRequest(username, repoName string, lastCommit time.Time) *http.Request {
	lastCommit.Add(time.Hour * -8)
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+username+"/"+repoName+"/commits?since="+lastCommit.Format("2006-01-02T15:01:05"), nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return req
}

func GetGitHubAuthReq(username, token string) *http.Request {
	head := username + token
	base64Head := util.EncodeBase64(head)
	fmt.Println(base64Head)
	req := &http.Request{}
	req.Header.Set("Authorization", "Basic "+base64Head)
	return req
}

func GetGitHubAuthReqWithToken(username, token string) *http.Request {
	req, err := http.NewRequest("GET", "https://api.github.com/users/"+username, nil)
	if err != nil {
		logger.Logger.Error(err)
	}
	req.Header.Set("Authorization", "token "+token)
	return req
}

func GetGitHubCreateRepoWithTemplateReq(token, name, tempOwner, tempName string) *http.Request {
	body := bytes.NewBuffer([]byte(fmt.Sprintf(`{
	"name": "%s"
	}`, name)))
	req, err := http.NewRequest("POST", "https://api.github.com/repos/"+tempOwner+"/"+tempName+"/generate", body)
	if err != nil {
		logger.Logger.Error(err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return req
}
