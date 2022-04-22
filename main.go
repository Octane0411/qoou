package main

import (
	"github.com/Octane0411/qoou/server/router"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(8)
	r := router.NewRouter()
	r.Run(":8080")
	//download.DownloadRepo("octane0411", "go-demo-1")
	//docker.DeployWithDockerFile("octane0411", "go-demo")
	//docker.StartFromImage("octane0411", "go-demo")
	//resp, err := v1.Cli.Do(v1.GetGitHubAuthReqWithToken("octane0411", "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"))
	/*	resp, err := v1.Cli.Do(v1.GetGitHubCreateRepoWithTemplateReq("11", "gd-3", "octane0411", "go-demo"))
		if err != nil {
			panic(err)
		}
		b := resp.Body
		all, _ := io.ReadAll(b)
		fmt.Println(string(all))*/
	//download.DownloadRepo("YanQiaoQi", "test")
	//project := &model.Project{
	//	Username: "YanQiaoQi",
	//	RepoName: "test",
	//	Template: "golang1.18",
	//}
	//err := docker.GenerateDockerfile(project)
	//if err != nil {
	//	panic(err)
	//}
}
