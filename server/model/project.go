package model

import (
	"gorm.io/gorm"
	"time"
)

type Project struct {
	Username    string    `json:"username"`
	RepoName    string    `json:"repoName"`
	Template    string    `json:"template"`
	Status      string    `json:"status"`
	Address     string    `json:"address"`
	LastCommit  time.Time `json:"lastCommit"`
	LastDeploy  time.Time `json:"lastDeploy"`
	ContainerID string    `json:"containerID" gorm:"column:container_id"`
	ImageID     string    `json:"imageID" gorm:"column:image_id"`
	*gorm.Model
}
