package service

import (
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
)

func CreateRepoFromTemplate(username, templateOwner, templateRepo string) {

}

func GetTokenByUsername(username string) string {
	token, err := dao.GetToken(username)
	if err != nil {
		logger.Logger.Error(err)
		return ""
	}
	return token
}
