package router

import (
	"github.com/Octane0411/qoou/server/middleware"
	v1 "github.com/Octane0411/qoou/server/router/api/v1"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	apiv1 := r.Group("/api/v1")
	apiv1.Use(middleware.JWT())
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

	authapi := r.Group("/api/auth")
	{
		authapi.POST("/login", v1.Login)
		authapi.POST("/register", v1.Register)
		authapi.POST("/captcha", v1.Captcha)
	}

	r.GET("/stream", func(c *gin.Context) {
		chanStream := make(chan int, 10)
		go func() {
			defer close(chanStream)
			for i := 0; i < 5; i++ {
				chanStream <- i
				time.Sleep(time.Second * 1)
			}
		}()
		go func(chanStream chan int) {
			c.Stream(func(w io.Writer) bool {
				if msg, ok := <-chanStream; ok {
					c.SSEvent("message", msg)
					return true
				}
				return false
			})
		}(chanStream)
	})

	return r
}
