package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/download"
	"github.com/Octane0411/qoou/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var ctx = context.Background()
var cli *client.Client
var logsDuration = 5 * time.Second

type ImageBuildResponse struct {
	Aux struct {
		ID string `json:"ID"`
	} `json:"aux"`
}

func init() {
	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func PullImage(imageName string) {
	//imageName = "golang:1.18"
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		logger.Logger.Error(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
}

func StartContainer1(username, repoName string) {
	cID, ok := GetContainerID(username, repoName)
	if !ok {
		logger.Logger.Error("容器不存在：", username+":"+repoName)
	}

	if err := cli.ContainerStart(ctx, cID, types.ContainerStartOptions{}); err != nil {
		logger.Logger.Error(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, cID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			logger.Logger.Error(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, cID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		logger.Logger.Error(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func GetContainerID(username, repoName string) (string, bool) {
	var cID string
	containers, _ := cli.ContainerList(ctx, types.ContainerListOptions{
		Quiet:   false,
		Size:    false,
		All:     true,
		Latest:  false,
		Since:   "",
		Before:  "",
		Limit:   0,
		Filters: filters.Args{},
	})
	for _, container := range containers {
		if container.Names[0] == "/"+username+"-"+repoName {
			cID = container.ID
		}
	}
	if cID == "" {
		return "", false
	}
	return cID, true
}
func GetImageID(username, repoName string) (string, bool) {
	var imageID string
	images, _ := cli.ImageList(ctx, types.ImageListOptions{})
	for _, image := range images {
		if image.RepoTags[0] == username+":"+repoName {
			imageID = image.ID
		}
	}
	if imageID == "" {
		return "", false
	}
	return imageID, true
}

func CopyContainerLogs(cID string) {
	//每隔一段时间读，并且从上次之后开始读
	beginTime := time.Now()
	//.Format("2006-01-02T15:04:05Z")
	sinceTime := beginTime
	for {
		since := sinceTime.Format("2006-01-02T15:04:05")
		time.Sleep(logsDuration)
		copyContainerLogs(cID, since)
		fmt.Println(time.Now().Format("2006-01-02T15:04:05"), " show since:", sinceTime.Format("2006-01-02T15:04:05"))
		sinceTime = sinceTime.Add(logsDuration)
	}
}

func copyContainerLogs(cID string, since string) {
	out, err := cli.ContainerLogs(ctx, cID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      since,
	})
	if err != nil {
		logger.Logger.Error(err)
	}
	//io.Copy(os.Stdout, out)
	stdcopy.StdCopy(os.Stdout, os.Stdin, out)
}

func CreateLogDir(username, repoName string) {

}

func CreateLogFile() {

}

func CreateImageWithDockerfile(username, repoName string) string {
	f, err := util.NewTarArchiveFromPath(download.GetRepoDir(username, repoName))
	if err != nil {
		logger.Logger.Error(err)
	}
	projectName := GetImageName(username, repoName)
	imageList, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		logger.Logger.Error(err)
	}

	for _, image := range imageList {
		if len(image.RepoTags) < 1 {
			continue
		}
		if image.RepoTags[0] == projectName {
			if _, err := cli.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{}); err != nil {
				logger.Logger.Error(err)
			}
		}
	}

	resp, err := cli.ImageBuild(ctx, f, types.ImageBuildOptions{
		Tags:        []string{username + "-" + repoName + ":latest"},
		NetworkMode: "host",
		Dockerfile:  "Dockerfile",
	})
	if err != nil {
		logger.Logger.Error(err)
	}
	var imageID string
	imageBuildResponse := &ImageBuildResponse{}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if strings.EqualFold(string(line), "\n") {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Logger.Error(err)
		}
		err = json.Unmarshal(line, imageBuildResponse)
		if err == nil {
			imageID = imageBuildResponse.Aux.ID
		}
	}
	//err = cli.ImageTag(ctx, imageID, username+"-"+repoName+":latest")
	if err != nil {
		logger.Logger.Error(err)
	}
	return imageID
}
func CreateAndStartContainer(username, repoName string) {
	imageName := GetImageName(username, repoName)
	var resp, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{"8080/tcp": []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: "8000",
		}}},
	}, nil, nil, username+"-"+repoName)
	cID := resp.ID
	if err := cli.ContainerStart(ctx, cID, types.ContainerStartOptions{}); err != nil {
		logger.Logger.Error(err)
	}
	if err != nil {
		logger.Logger.Error(err)
	}
}

func StartContainer(username, repoName string) {
	cID, ok := GetContainerID(username, repoName)
	if !ok {
		return
	}
	if err := cli.ContainerStart(ctx, cID, types.ContainerStartOptions{}); err != nil {
		logger.Logger.Error(err)
	}
}

func DockerDaemonAlive() bool {
	_, err := cli.Ping(ctx)
	return err == nil
}

func GetImageName(username string, repoName string) string {
	return username + "-" + repoName + ":" + "latest"
}

// deprecated, use CreateImageWithDockerfile
func DeployGoProjectToContainer(username, repoName string) {
	cID, ok := GetContainerID(username, repoName)
	if ok {
		err := cli.ContainerStop(ctx, cID, nil)
		err = cli.ContainerRemove(ctx, cID, types.ContainerRemoveOptions{})
		if err != nil {
			log.Fatalln("err in remove container: ", err)
		}
	}

	var resp, err = cli.ContainerCreate(ctx, &container.Config{

		Cmd:          []string{"go", "run", "/home/" + repoName + "/main.go"},
		Image:        "golang:1.18",
		ExposedPorts: nat.PortSet{"9000": struct{}{}},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{"9000": []nat.PortBinding{nat.PortBinding{
			HostIP:   "127.0.0.1",
			HostPort: "9000",
		}}},
	}, nil, nil, username+"-"+repoName)

	fmt.Printf("starting container: %v\n", resp.ID)

	if err != nil {
		logger.Logger.Error(err)
	}
	f, err := util.NewTarArchiveFromPath(download.GetRepoDir(username, repoName))
	if err != nil {
		log.Fatalln(err)
	}
	cli.CopyToContainer(ctx, resp.ID, "/home", f, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
		CopyUIDGID:                false,
	})

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		logger.Logger.Error(err)
	}

	go CopyContainerLogs(resp.ID)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			logger.Logger.Error(err)
		}
	case <-statusCh:
	}

	//cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

}