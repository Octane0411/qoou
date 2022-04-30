package model

import "gorm.io/gorm"

type User struct {
	Username string
	Token    string
	Email    string
	Password string
	*gorm.Model
}

func (user User) TableName() string {
	return "users"
}
