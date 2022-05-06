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
	"sync"
	"time"
)

var host = "http://localhost:8080/preview"

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
		err = docker.StartContainer(request.Username, request.RepoName)
		if err != nil {
			err = docker.GenerateDockerfile(project)
			if err != nil {
				c.JSON(500, gin.H{"msg": "error"})
			}
			imageID = docker.CreateImageWithDockerfile(request.Username, request.RepoName)
			containerID, err = docker.CreateContainer(request.Username, request.RepoName)
			err = docker.StartContainer(request.Username, request.RepoName)
		}
	} else {
		lastCommit := gResp[0].Commit.Committer.Date
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
		containerID, err = docker.CreateContainer(request.Username, request.RepoName)
		err = docker.StartContainer(request.Username, request.RepoName)
	}
	// write db lastDeploy
	project.Address = host + "/" + project.Username + "/" + project.RepoName + "/"
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
	if username == "" {
		c.JSON(400, gin.H{
			"message": "invalid request",
		})
		return
	}
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

	project, err := dao.GetProject(username, repoName)
	if err != nil {
		c.JSON(500, gin.H{"msg": "db error"})
		return
	}
	project.LastRun = time.Now()
	project.Status = "running"
	err = dao.UpdateProject(project)
	if err != nil {
		c.JSON(500, gin.H{"msg": "db error"})
		return
	}

	err = docker.StartContainer(username, repoName)
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

	project, err := dao.GetProject(username, repoName)
	if err != nil {
		c.JSON(500, gin.H{"msg": "db error"})
		return
	}
	project.LastRun = time.Now()
	project.Status = "stop"
	err = dao.UpdateProject(project)
	if err != nil {
		c.JSON(500, gin.H{"msg": "db error"})
		return
	}

	err = docker.StopContainer(username, repoName)
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

	done := make(chan struct{}, 1)
	defer close(done)

	//reader := bufio.NewReader(logsReader)
	scanner := bufio.NewScanner(logsReader)
	scanner.Split(bufio.ScanLines)

	wg := sync.WaitGroup{}
	wg.Add(2)
	// 写goroutine
	go func() {
		output := make(chan interface{})
		//单独开启一个goroutine，因为scanner.Scan()会阻塞，不能放在for select的default中，否则会一直阻塞
		go func() {
			for scanner.Scan() {
				output <- scanner.Bytes()
			}
			output <- scanner.Err()
		}()
		for {
			select {
			case <-done:
				wg.Done()
				return
			case lineOrErr := <-output:
				if line, ok := lineOrErr.([]byte); ok {
					// 向websocket写入
					err = c.WriteMessage(1, line)
					if err != nil {
						logger.Logger.Error("write: ", err)
						break
					}
				} else {
					err, ok := lineOrErr.(error)
					if ok {
						logger.Logger.Error(err)
					}
				}
			}
		}
	}()
	// 读goroutine
	go func() {
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				logger.Logger.Error(err)
				break
			}
			if string(data) == "end" {
				err := c.Close()
				if err != nil {
					logger.Logger.Error(err)
				}
				done <- struct{}{}
				wg.Done()
				logsReader.Close()
				return
			}
		}
	}()
	wg.Wait()
	logger.Logger.Info("doneeeeee")
}
