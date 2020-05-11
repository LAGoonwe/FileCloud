package handler

import (
	"FileCloud/util"
	"net/http"
)

/**
处理手机和邮箱验证的handler
*/

var CODE string

/**
邮件验证处理器
*/
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
	w.Write([]byte(CODE))
}

/**
手机验证处理器
*/

func PhoneValideHandler(w http.ResponseWriter, r *http.Request) {
	var CODE = util.CreateCaptcha()

	// 接收手机号参数
	r.ParseForm()
	phoneVal := r.Form.Get("phoneVal")

	err := util.SendPhoneCode(CODE, phoneVal)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(CODE))
}
