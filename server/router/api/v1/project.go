package v1

import (
	"bufio"
	"encoding/json"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/server/docker"
	"github.com/Octane0411/qoou/server/download"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	var imageID, containerID string
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
		err = docker.StartContainer(request.Username, request.RepoName)
		if err != nil {
			err = docker.GenerateDockerfile(project)
			if err != nil {
				c.JSON(500, gin.H{"msg": "error"})
			}
			imageID = docker.CreateImageWithDockerfile(request.Username, request.RepoName)
			containerID = docker.CreateAndStartContainer(request.Username, request.RepoName)
		}
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
		err = docker.GenerateDockerfile(project)
		if err != nil {
			c.JSON(500, gin.H{"msg": "error"})
		}
		imageID = docker.CreateImageWithDockerfile(request.Username, request.RepoName)
		containerID = docker.CreateAndStartContainer(request.Username, request.RepoName)

	}
	// write db lastDeploy
	project.LastRun = time.Now()
	project.ImageID = imageID
	project.ContainerID = containerID
	project.Status = "running"
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

func GetProjectsByUsername(c *gin.Context) {
	username := c.Query("username")
	repoName := c.Query("repoName")
	if repoName != "" {
		project, err := dao.GetProject(username, repoName)
		if err != nil {
			c.JSON(500, gin.H{"msg": "db error"})
			return
		}
		c.JSON(200, gin.H{"msg": "success", "data": project})
		return
	}
	projects, err := dao.GetProjectByUsername(username)
	if err != nil {
		c.JSON(500, gin.H{"msg": "db error"})
		return
	}
	c.JSON(200, gin.H{"msg": "success", "data": projects})
}

func StartContainer(c *gin.Context) {
	username := c.Query("username")
	repoName := c.Query("repoName")
	err := docker.StartContainer(username, repoName)
	if err != nil {
		logger.Logger.Error("error start container:", err)
		c.JSON(500, gin.H{"msg": "error"})
		return
	}
	c.JSON(200, gin.H{"msg": "success"})
}

func StopContainer(c *gin.Context) {
	username := c.Query("username")
	repoName := c.Query("repoName")
	err := docker.StopContainer(username, repoName)
	if err != nil {
		logger.Logger.Error("error start container:", err)
		c.JSON(500, gin.H{"msg": "error"})
		return
	}
	c.JSON(200, gin.H{"msg": "success"})
}

func GetLog(c *gin.Context) {
	username := c.Query("username")
	repoName := c.Query("repoName")
	getLog(c.Writer, c.Request, username, repoName)
}

func getLog(w http.ResponseWriter, r *http.Request, username, repoName string) {
	cID, ok := docker.GetContainerID(username, repoName)
	if !ok {
		logger.Logger.Error("容器不存在")
	}
	logsReader := docker.ContainerLogs(cID)
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Error("upgrade:", err)
		return
	}
	defer c.Close()
	//ticker := time.NewTicker(1 * time.Second)
	//defer ticker.Stop()
	reader := bufio.NewReader(logsReader)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			logger.Logger.Error(err)
			break
		}
		err = c.WriteMessage(1, line)
		if err != nil {
			logger.Logger.Error("write: ", err)
			break
		}

	}
}

func GetGitHubCommitsRequest(username, repoName string, lastCommit time.Time) *http.Request {
	lastCommit.Add(time.Hour * -8)
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+username+"/"+repoName+"/commits?since="+lastCommit.Format("2006-01-02T15:01:05"), nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return req
}
