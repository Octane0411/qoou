package docker

import (
	"github.com/Octane0411/qoou/common/logger"
	"github.com/docker/docker/api/types"
	"io"
)

func ContainerLogs(cID string) io.ReadCloser {
	logsReader, err := cli.ContainerLogs(ctx, cID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		//Tail:       "50",
	})
	if err != nil {
		logger.Logger.Error(err)
	}
	return logsReader
}
