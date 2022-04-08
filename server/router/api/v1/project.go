package v1

import (
	"encoding/json"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/server/docker"
	"github.com/Octane0411/qoou/server/download"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

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
type GitHubCommitResponse []GitHubCommit

func Deploy(c *gin.Context) {
	// require: username, repoName
	var request DeployRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "invalid request",
		})
		return
	}
	project, err := dao.GetProject(request.Username, request.RepoName)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "project doesn't exist",
		})
		return
	}
	// check commit
	resp, err := Cli.Do(GetGitHubCommitsRequest(request.Username, request.RepoName, project.LastCommit))
	if err != nil {
		logger.Logger.Error(err)
		c.JSON(400, gin.H{
			"message": "failed to get commit",
		})
		return
	}
	b, _ := io.ReadAll(resp.Body)
	gResp := GitHubCommitResponse{}
	err = json.Unmarshal(b, &gResp)
	logger.Logger.Infoln(string(b))
	if err != nil {
		logger.Logger.Error(err)
		c.JSON(500, gin.H{
			"message": "GitHub request error",
		})
	}
	if len(gResp) < 1 {
		// no update
		logger.Logger.Info("no update")
		//TODO:检查是否有异常，比如有image却没有container
		docker.StartContainer(request.Username, request.RepoName)
	} else {
		lastCommit := gResp[0].Commit.Committer.Date
		// update

		// write db
		project.LastCommit = lastCommit
		err = dao.UpdateProject(project)
		if err != nil {
			logger.Logger.Error(err)
			c.JSON(500, gin.H{
				"message": "write db error",
			})
			return
		}
		// Download latest project
		download.DownloadRepo(request.Username, request.RepoName)
		imageID := docker.CreateImageWithDockerfile(request.Username, request.RepoName)
		containerID := docker.CreateAndStartContainer(request.Username, request.RepoName)
		project.ImageID = imageID
		project.ContainerID = containerID
		err = dao.UpdateProject(project)
		if err != nil {
			logger.Logger.Error(err)
			c.JSON(500, gin.H{
				"message": "write db error",
			})
			return
		}
	}
	// write db lastDeploy
	project.LastRun = time.Now()
	err = dao.UpdateProject(project)
	if err != nil {
		logger.Logger.Error(err)
		c.JSON(500, gin.H{
			"message": "write db error",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success deploy",
	})
}

func GetGitHubCommitsRequest(username, repoName string, lastCommit time.Time) *http.Request {
	lastCommit.Add(time.Hour * -8)
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+username+"/"+repoName+"/commits?since="+lastCommit.Format("2006-01-02T15:01:05"), nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return req
}
