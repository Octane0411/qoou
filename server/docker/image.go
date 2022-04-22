package docker

import (
	"bufio"
	"encoding/json"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/download"
	"github.com/Octane0411/qoou/util"
	"github.com/docker/docker/api/types"
	"io"
	"os"
	"strings"
)

type ImageBuildResponse struct {
	Aux struct {
		ID string `json:"ID"`
	} `json:"aux"`
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

func GetImageName(username string, repoName string) string {
	return username + "-" + repoName + ":" + "latest"
}
