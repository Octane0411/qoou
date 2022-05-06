package v1

import (
	"github.com/Octane0411/qoou/common/email"
	"github.com/Octane0411/qoou/common/logger"
	"github.com/Octane0411/qoou/common/qoou_jwt"
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/server/model"
	"github.com/Octane0411/qoou/server/service"
	"github.com/gin-gonic/gin"
)

func Captcha(c *gin.Context) {
	json := make(map[string]any)
	err := c.BindJSON(&json)
	email1 := json["email"].(string)
	err = service.SendCaptchaEmail(email1)
	if err != nil {
		logger.Logger.Error(err)
		c.JSON(500, gin.H{"message": "send email error"})
		return
	}
	c.JSON(200, gin.H{"message": "send email success"})
}

func Register(c *gin.Context) {
	json := make(map[string]any)
	c.BindJSON(&json)
	email1 := json["email"].(string)
	username := json["username"].(string)
	password := json["password"].(string)
	captcha := json["captcha"].(string)
	validCaptcha := service.ValidCaptcha(email1, captcha)
	if !validCaptcha {
		c.JSON(403, gin.H{"message": "验证码错误"})
		return
	}
	userByUsername := dao.GetUserByUsername(username)
	if userByUsername != nil {
		c.JSON(403, gin.H{"message": "用户名已存在"})
		return
	}
	err := dao.CreateUser(&model.User{
		Username: username,
		Email:    email1,
		Password: password,
	})
	if err != nil {
		logger.Logger.Error(err)
		c.JSON(500, gin.H{"message": "create user error"})
		return
	}
	c.JSON(200, gin.H{"message": "注册成功"})
}

func Login(c *gin.Context) {
	json := make(map[string]any)
	c.BindJSON(&json)
	usernameOrEmail := json["usernameOrEmail"].(string)
	password := json["password"].(string)
	var user *model.User
	if email.ValidEmail(usernameOrEmail) {
		user = dao.GetUserByEmail(usernameOrEmail)
	} else {
		user = dao.GetUserByUsername(usernameOrEmail)
	}
	if user == nil {
		c.JSON(403, gin.H{"message": "用户名或邮箱不存在"})
		return
	}
	// TODO: 密码加密
	if user.Password != password {
		c.JSON(403, gin.H{"message": "密码错误"})
		return
	}
	token, err := qoou_jwt.CreateToken(user.Username)
	if err != nil {
		c.JSON(500, gin.H{"message": "create token error"})
		return
	}
	c.JSON(200, gin.H{"token": token, "msg": "登录成功"})
}
