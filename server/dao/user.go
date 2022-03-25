package dao

import (
	"github.com/Octane0411/qoou/common/db"
	"github.com/Octane0411/qoou/server/model"
)

func FindUser(username string) bool {
	user := &model.User{}
	db.DB.Where("username = ?", username).First(&user)
	if user.Username == "" {
		return false
	}
	return true
}
