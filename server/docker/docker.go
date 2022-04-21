package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Octane0411/qoou/common/global"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/server/download"
	"github.com/Octane0411/qoou/server/model"
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
	"strconv"
	"strings"
	"text/template"
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

func ContainerLogs(cID string) io.ReadCloser {
	logsReader, err := cli.ContainerLogs(ctx, cID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "50",
	})
	if err != nil {
		logger.Logger.Error(err)
	}
	return logsReader
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
			logger.Logger.Info("imageID:", imageID)
		}
	}
	//err = cli.ImageTag(ctx, imageID, username+"-"+repoName+":latest")
	if err != nil {
		logger.Logger.Error(err)
	}
	return imageID
}
func CreateAndStartContainer(username, repoName string) string {
	imageName := GetImageName(username, repoName)
	// generate a free port
	freePort, err := util.GetFreePort()
	if err != nil {
		logger.Logger.Error(err)
	}
	port := strconv.Itoa(freePort)
	// write to redis
	err = dao.SetPort(port, username, repoName)
	if err != nil {
		logger.Logger.Error(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{"8080/tcp": []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: port,
		}}},
	}, nil, nil, username+"-"+repoName)
	if err != nil {
		logger.Logger.Error("image create:", err)
	}

	cID := resp.ID
	if err = cli.ContainerStart(ctx, cID, types.ContainerStartOptions{}); err != nil {
		logger.Logger.Error(err)
	}
	return cID
}

func StartContainer(username, repoName string) error {
	cID, ok := GetContainerID(username, repoName)
	if !ok {
		return fmt.Errorf("容器不存在")
	}
	err := cli.ContainerStart(ctx, cID, types.ContainerStartOptions{})
	if err != nil {
		logger.Logger.Error(err)
	}
	return err
}

func StopContainer(username, repoName string) error {
	cID, ok := GetContainerID(username, repoName)
	if !ok {
		return fmt.Errorf("容器不存在")
	}
	err := cli.ContainerStop(ctx, cID, nil)
	if err != nil {
		logger.Logger.Error(err)
	}
	return err
}

func DockerDaemonAlive() bool {
	_, err := cli.Ping(ctx)
	return err == nil
}

func GetImageName(username string, repoName string) string {
	return username + "-" + repoName + ":" + "latest"
}

func GenerateDockerfile(project *model.Project) error {
	projectDir := download.GetRepoDir(project.Username, project.RepoName)
	dockerfileDir := projectDir + "/Dockerfile"
	ok := FileExsits(dockerfileDir)
	if ok {
		//Dockerfile存在

	} else {
		//Dockerfile不存在
		file, err := CreateFile(dockerfileDir)
		if err != nil {
			return err
		}
		dockerfileTemplate, err := GetDockerfileTemplate(project.Template)
		if err != nil {
			return err
		}
		dockerfileTemplate.Execute(file, project)
	}
	return nil
}

func CreateFile(dir string) (*os.File, error) {
	//file, err := os.OpenFile(dir, os.O_CREATE, os.ModePerm)
	f, err := os.Create(dir)
	if err != nil {
		logger.Logger.Error("error on create file", err)
		return nil, err
	}
	return f, nil
}

func FileExsits(dir string) bool {
	if _, err := os.Stat(dir); err == nil {
		// path/to/whatever exists
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		panic(err)
	}
}

func GetDockerfileTemplate(templ string) (*template.Template, error) {
	t, err := template.New(templ).Parse(global.TemplateDockerfilemap[templ])
	if err != nil {
		return nil, err
	}
	return t, nil
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

// deprecated
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
