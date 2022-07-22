package docker

import (
	"fmt"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	"strconv"
)

func CreateAndStartContainer(username, repoName string) (string, string) {
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
	return cID, port
}
func CreateContainer(username, repoName string) (string, error) {
	imageName := GetImageName(username, repoName)
	freePort, err := util.GetFreePort()
	if err != nil {
		logger.Logger.Error(err)
	}
	port := strconv.Itoa(freePort)
	logger.Logger.Info(port)
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
	return cID, nil
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

func RemoveContainer(cID string) error {
	err := cli.ContainerRemove(ctx, cID, types.ContainerRemoveOptions{})
	return err
}
