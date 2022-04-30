package v1

import (
	"github.com/Octane0411/qoou/server/dao"
	"github.com/Octane0411/qoou/server/model"
	"github.com/Octane0411/qoou/server/service"
	"github.com/gin-gonic/gin"
)

func Captcha(c *gin.Context) {
	email := c.PostForm("email")
	err := service.SendCaptchaEmail(email)
	if err != nil {
		c.JSON(500, gin.H{"message": "send email error"})
		return
	}
}

func Register(c *gin.Context) {
	email := c.PostForm("email")
	username := c.PostForm("username")
	password := c.PostForm("password")
	captcha := c.PostForm("captcha")
	validCaptcha := service.ValidCaptcha(email, captcha)
	if !validCaptcha {
		c.JSON(200, gin.H{"message": "验证码错误"})
		return
	}
	// TODO: 写入数据库
	dao.Create(&model.User{
		Username: username,
		Email:    email,
		Password: password,
	})
	c.JSON(200, gin.H{"message": "注册成功"})
}

func Login(c *gin.Context) {

}
