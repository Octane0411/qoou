package email

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"log"
	"net/mail"
	"strings"
)

// 全局变量，因为发件人账号、密码，需要在发送时才指定
var serverHost, fromEmail, fromPasswd string
var serverPort int
var m *gomail.Message

// SendEmail 发送邮件
// Toers: 接受者邮件，如有多个，用逗号分割，不能为空
// Ccers: 抄送者邮件，如有多个，用逗号分割，可以为空
// subject: 邮件的主题
// body: 邮件的内容
func SendEmail(Toers, Ccers, subject, body, name string) error {
	toers := []string{}

	//serverHost = config.EmailSetting.ServerHost
	//serverPort = config.EmailSetting.ServerPort
	//fromEmail = config.EmailSetting.FromEmail
	//fromPasswd = config.EmailSetting.FromPassword
	serverHost = "smtp.yeah.net"
	serverPort = 465
	fromEmail = "qoou_no_reply@yeah.net"
	fromPasswd = "OHOPJTVNYUOBFJHH"
	m = gomail.NewMessage()

	if len(Toers) == 0 {
		return fmt.Errorf("Toers不能为空")
	} else {
		for _, tmp := range strings.Split(Toers, ",") {
			if !ValidEmail(tmp) {
				return fmt.Errorf("邮箱格式不正确：%v", tmp)
			}
			toers = append(toers, strings.TrimSpace(tmp))
		}
	}
	// 收件人可以有多个，故用此方式
	m.SetHeader("To", toers...)
	//抄送列表
	if len(Ccers) != 0 {
		for _, tmp := range strings.Split(Ccers, ",") {
			if !ValidEmail(tmp) {
				return fmt.Errorf("邮箱格式不正确：%v", tmp)
			}
			toers = append(toers, strings.TrimSpace(tmp))
		}
		m.SetHeader("Cc", toers...)
	}
	// 发件人
	m.SetAddressHeader("From", fromEmail, name)
	// 主题
	m.SetHeader("Subject", subject)
	// 正文
	m.SetBody("text/html", body)
	d := gomail.NewDialer(serverHost, serverPort, fromEmail, fromPasswd)
	// 发送
	err := d.DialAndSend(m)
	if err != nil {
		log.Printf("sendEmail DialAndSend error: %v", err)
		return err
	}
	return nil
}

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
