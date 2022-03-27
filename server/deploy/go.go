package deploy

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/Octane0411/qoou/server/download"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ctx = context.Background()
var cli *client.Client
var forceStopDuration = (1 * time.Second)

func init() {
	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func PullGoImage() {
	imageName := "golang:1.18"
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
}

func DeployGoProjectToContainer(username, repoName string) {
	cID, ok := GetContainerIDByName(username, repoName)
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

	fmt.Println(resp.ID)

	if err != nil {
		panic(err)
	}
	f, err := NewTarArchiveFromPath(download.GetRepoDir(username, repoName))
	if err != nil {
		log.Fatalln(err)
	}
	cli.CopyToContainer(ctx, resp.ID, "/home", f, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
		CopyUIDGID:                false,
	})

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	//cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

}

func StartGoPrejectContainer(username, repoName string) {
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
		log.Fatalln("没有找到这个容器")
	}

	if err := cli.ContainerStart(ctx, cID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, cID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, cID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func NewTarArchiveFromPath(path string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	ok := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(strings.Replace(file, path, "", -1), string(filepath.Separator))
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}
		return nil
	})

	if ok != nil {
		return nil, ok
	}
	ok = tw.Close()
	if ok != nil {
		return nil, ok
	}
	return bufio.NewReader(&buf), nil
}

func GetContainerIDByName(username, repoName string) (string, bool) {
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
