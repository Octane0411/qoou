package dao

import (
	"github.com/Octane0411/qoou/common/db"
	"github.com/Octane0411/qoou/server/model"
)

func FindUser(username string) bool {
	user := model.NewUser()
	db.DB.Where("username = ?", username).First(&user)
	if user.Username == "" {
		return false
	}
	return true
}

func GetToken(username string) (string, error) {
	user := model.NewUser()
	err := db.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return "", err
	}
	return user.Token, nil
}

func CreateUser(user *model.User) error {
	db.DB.Create(user)
	return nil
}

func UpdateUser(user *model.User) error {
	db.DB.Model(&user).Updates(user)
	return nil
}

func GetUserByEmail(email string) *model.User {
	user := model.NewUser()
	db.DB.Where("email = ?", email).First(&user)
	return user
}

func GetUserByUsername(username string) *model.User {
	user := model.NewUser()
	db.DB.Where("username = ?", username).First(&user)
	return user
}
