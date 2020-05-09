package util

/**
邮箱验证工具类，使用Gomail发送电子邮件
*/

import (
	"fmt"
	"github.com/gomail"
)

const (
	// 邮件服务器地址
	MAIL_HOST = "smtp.163.com"
	// 端口
	MAIL_PORT = 465
	// 发送邮件用户账号
	MAIL_USER = "14779873246@163.com"
	// 授权密码
	MAIL_PWD = "lwq121797"
)

/*
title 使用gomail发送邮件
@param []string mailAddress 收件人邮箱
@param string subject 邮件主题
@param string body 邮件内容
@return error
*/
func SendGoMail(mailAddress []string, subject string, body string) error {
	m := gomail.NewMessage()
	// 这种方式可以添加别名，即 nickname， 也可以直接用<code>m.SetHeader("From", MAIL_USER)</code>
	nickname := "ZeusCloud"
	m.SetHeader("From", nickname+"<"+MAIL_USER+">")
	// 发送给多个用户
	m.SetHeader("To", mailAddress...)
	// 设置邮件主题
	m.SetHeader("Subject", subject)
	// 设置邮件正文
	m.SetBody("text/html", body)
	d := gomail.NewDialer(MAIL_HOST, MAIL_PORT, MAIL_USER, MAIL_PWD)
	// 发送邮件
	err := d.DialAndSend(m)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func main2() {
	EmailTitle := "宙斯云盘注册"
	EmailBody := "【宙斯云盘】您的验证码是" +
		"<p style='color:red'>" + CreateCaptcha() + "</p>" +
		"如非本人操作请忽略该邮件"
	SendGoMail([]string{"736750759@qq.com"}, EmailTitle, EmailBody)
}
