package handler

import (
	"FileCloud/util"
	"net/http"
	"time"
)

/**
处理手机和邮箱验证的handler
*/

var CODE string

func EmailValideHandler(w http.ResponseWriter, r *http.Request) {
	// 设定邮件头和邮件体

	CODE = util.CreateCaptcha()

	EmailTitle := "宙斯云盘注册"
	EmailBody := "【宙斯云盘】您的验证码是" +
		"<p style='color:red'>" + CODE + "</p>" +
		"如非本人操作请忽略该邮件"

	// 接收邮箱参数
	r.ParseForm()
	emailVal := r.Form.Get("emailVal")
	err := util.SendGoMail([]string{emailVal}, EmailTitle, EmailBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	time.Sleep(time.Duration(60) * time.Second)
	w.Write([]byte(CODE))
}
