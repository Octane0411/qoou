package router

import (
	"github.com/Octane0411/qoou/server/middleware"
	v1 "github.com/Octane0411/qoou/server/router/api/v1"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	apiv1 := r.Group("/api/v1")

	{
		apiv1.POST("/github_token", v1.GitHubToken)

		apiv1.POST("/repo", v1.CreateRepoWithTemplate)

		apiv1.POST("/project/deploy", v1.Deploy)
		apiv1.GET("/project/", v1.GetProjectsByUsername)
		apiv1.GET("/project/start", v1.StartContainer)
		apiv1.GET("/project/stop", v1.StopContainer)
		apiv1.GET("/project/preview/:username/:repoName/*all", v1.Forward)

		apiv1.GET("/project/log", v1.GetLog)

	}

	return r
}
