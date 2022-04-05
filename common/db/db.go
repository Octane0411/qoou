package db

import (
	"context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB = NewDbEngine()
var ctx = context.Background()

func NewDbEngine() *gorm.DB {
	// data source name
	dsn := "root:123456@tcp(127.0.0.1:3305)/qoou?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	return db
}
