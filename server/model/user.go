package model

import "gorm.io/gorm"

type User struct {
	Username string
	Token    string
	*gorm.Model
}
