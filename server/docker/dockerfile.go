package docker

import (
	"errors"
	"github.com/Octane0411/qoou/common/global"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/download"
	"github.com/Octane0411/qoou/server/model"
	"os"
	"text/template"
)

func GenerateDockerfile(project *model.Project) error {
	projectDir := download.GetRepoDir(project.Username, project.RepoName)
	dockerfileDir := projectDir + "/Dockerfile"
	ok := fileExsits(dockerfileDir)
	if ok {
		//Dockerfile存在

	} else {
		//Dockerfile不存在
		file, err := createFile(dockerfileDir)
		if err != nil {
			return err
		}
		dockerfileTemplate, err := GetDockerfileTemplate(project.Template)
		if err != nil {
			return err
		}
		dockerfileTemplate.Execute(file, project)
	}
	return nil
}

func createFile(dir string) (*os.File, error) {
	//file, err := os.OpenFile(dir, os.O_CREATE, os.ModePerm)
	f, err := os.Create(dir)
	if err != nil {
		logger.Logger.Error("error on create file", err)
		return nil, err
	}
	return f, nil
}

func fileExsits(dir string) bool {
	if _, err := os.Stat(dir); err == nil {
		// path/to/whatever exists
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		panic(err)
	}
}

func GetDockerfileTemplate(templ string) (*template.Template, error) {
	t, err := template.New(templ).Parse(global.TemplateDockerfileMap[templ])
	if err != nil {
		return nil, err
	}
	return t, nil
}
