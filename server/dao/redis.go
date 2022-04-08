package dao

import "github.com/Octane0411/qoou/common/rdb"

func SetPort(port string, username string, repoName string) error {
	err := rdb.RDB.HSet(rdb.Ctx, "portMap", username+":"+repoName, port).Err()
	return err
}

func GetPort(username string, repoName string) (string, error) {
	result, err := rdb.RDB.HGet(rdb.Ctx, "portMap", username+":"+repoName).Result()
	return result, err
}
