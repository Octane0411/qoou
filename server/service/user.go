package service

import (
	"fmt"
	email_ "github.com/Octane0411/qoou/common/email"
	"github.com/Octane0411/qoou/common/rdb"
	"github.com/Octane0411/qoou/util"
	"time"
)

var captchaExpireDuration = time.Minute * 10

func SendCaptchaEmail(email string) error {
	captcha := util.GetCaptcha()
	err := email_.SendEmail(email, "", "qoou", fmt.Sprintf("感谢你使用qoou，你的验证码是：%s", captcha), "qoou")
	if err != nil {
		return err
	}
	// 写入到缓存
	rdb.RDB4Server.Set(rdb.Ctx, email, captcha, captchaExpireDuration)
	return nil
}

func ValidCaptcha(email, captcha string) bool {
	captcha_ := rdb.RDB4Server.Get(rdb.Ctx, email).Val()
	if captcha == captcha_ {
		return true
	}
	return false
}
