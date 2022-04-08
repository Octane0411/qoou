package v1

import (
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ForwardHandler(writer http.ResponseWriter, request *http.Request) {
	port, err := dao.GetPort(request.FormValue("username"), request.FormValue("repoName"))
	if nil != err {
		logger.Logger.Error(err)
		return
	}
	// Get Port From Redis
	u, err := url.Parse("http://127.0.0.1:" + port + "/")
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
