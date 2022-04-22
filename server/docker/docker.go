package docker

import (
	"context"
	"fmt"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/download"
	"github.com/Octane0411/qoou/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"log"
	"os"
	"time"
)

var ctx = context.Background()
var cli *client.Client
var logsDuration = 5 * time.Second

func init() {
	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func DockerDaemonAlive() bool {
	_, err := cli.Ping(ctx)
	return err == nil
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

// deprecated
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

// deprecated
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
