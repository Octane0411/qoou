package v1

import (
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func Forward(c *gin.Context) {
	username := c.Param("username")
	repoName := c.Param("repoName")
	port, err := dao.GetPort(username, repoName)
	if nil != err {
		logger.Logger.Error(err)
		return
	}
	userURI := "/api/v1/project/preview/" + username + "/" + repoName + "/"
	trueURI := strings.TrimPrefix(c.Request.RequestURI, userURI)
	ForwardHandler(c.Writer, c.Request, port, trueURI)
}

func ForwardHandler(writer http.ResponseWriter, request *http.Request, port string, trueURI string) {
	urlStr := "http://127.0.0.1:" + port + trueURI
	logger.Logger.Info(urlStr)
	u, err := url.Parse(urlStr)
	if nil != err {
		logger.Logger.Error(err)
		return
	}

	proxy := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL = u
		},
	}

	proxy.ServeHTTP(writer, request)
}
