package dao

import (
	"github.com/Octane0411/qoou/common/db"
	"github.com/Octane0411/qoou/server/model"
)

func CreateProject(project *model.Project) error {
	err := db.DB.Create(project).Error
	if err != nil {
		return err
	}
	return nil
}

func GetProject(username, repoName string) (*model.Project, error) {
	project := &model.Project{}
	err := db.DB.Where("username = ? AND repo_name = ?", username, repoName).First(project).Error
	if err != nil {
		return nil, err
	}
	return project, nil
}

func UpdateProject(project *model.Project) error {
	err := db.DB.Updates(project).Error
	if err != nil {
		return err
	}
	return nil
}
