package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Octane0411/qoou/common/db"
	"github.com/Octane0411/qoou/common/global"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/server/model"
	"github.com/Octane0411/qoou/server/service"
	"github.com/Octane0411/qoou/util"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

// https://docs.github.com/cn/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
// 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
func LoginGithubWithToken(c *gin.Context) {
	tr := &TokenRequest{}
	c.BindJSON(tr)
	client := &http.Client{}
	requestBody := fmt.Sprintf(`{
	"name": "%s"
	}`, "gd-2")
	var jsonStr = []byte(requestBody)
	req, err := http.NewRequest("POST", "https://api.github.com/repos/octane0411/go-demo/generate", bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	//req.Header.Set("Authorization", "Basic "+base64Head)

	//req, err := http.NewRequest("GET", "https://api.github.com/users/"+tr.Username, nil)
	req.Header.Set("Authorization", "token "+tr.Token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	if err != nil {
		log.Fatal(err)
	}
}

func GitHubToken(c *gin.Context) {
	tr := &TokenRequest{}
	err := c.BindJSON(tr)
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "请求格式错误",
		})
		return
	}
	//验证是否注册过
	user := &model.User{
		Username: tr.Username,
		Token:    tr.Token,
	}
	ok := dao.FindUser(tr.Username)
	if ok {
		db.DB.Updates(user)
		c.JSON(200, gin.H{"msg": "token已更新"})
		return
	}
	//新增用户
	db.DB.Updates(user)
	c.JSON(200, gin.H{"msg": "success"})
}

func CreateRepoWithTemplate(c *gin.Context) {
	req := &CreateRepoWithTemplateRequest{}
	err := c.BindJSON(req)
	if err != nil {
		logger.Logger.Error(err)
	}
	template, ok := global.TemplateMap[req.Template]
	if !ok {
		c.JSON(200, gin.H{"msg": "模板不存在"})
		return
	}

	// TODO:封装到service中，并且另起一个协程
	// 写入数据库
	t := util.GetZeroTime()
	err = dao.CreateProject(&model.Project{
		Username:   req.Username,
		RepoName:   req.RepoName,
		Template:   req.Template,
		LastCommit: t,
		LastRun:    t,
	})
	if err != nil {
		logger.Logger.Error(err)
		c.JSON(500, gin.H{"msg": "写入数据库失败"})
		return
	}

	// 创建仓库
	gReq := GetGitHubCreateRepoWithTemplateReq(service.GetTokenByUsername(req.Username), req.RepoName, template[0], template[1])
	resp, err := Cli.Do(gReq)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error(err)
	}
	gResp := &GitHubErrorResponse{}
	json.Unmarshal(body, gResp)
	if gResp.Message != "" {
		logger.Logger.Debugln(gResp.Message)
		c.JSON(500, gin.H{"msg": gResp.Message})
		return
	}

	c.JSON(200, gin.H{"msg": "创建成功"})
}
