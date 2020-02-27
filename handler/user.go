package handler

import (
	dblayer "FileCloud/db"
	"FileCloud/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pwd_salt="*#890"
)


//处理用户注册请求
func SignupHandler(w http.ResponseWriter,r *http.Request)  {
	if r.Method == http.MethodGet {
		data,err := ioutil.ReadFile("src/FileCloud/static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(data)
		return
	}


	//表单参数解析
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username)<3 || len(passwd)<5{
		w.Write([]byte("invalid parameter"))
		return
	}

	//对用户密码进行哈希的加密处理
	enc_passwd := util.Sha1([]byte(passwd+pwd_salt))
	suc := dblayer.UserSignup(username,enc_passwd)
	if suc {
		w.Write([]byte("SUCCESS"))
	}else {
		w.Write([]byte("FAILED"))
	}
}

//用户登录接口
func SignInHandler(w http.ResponseWriter,r *http.Request){

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encpasswd := util.Sha1([]byte(password+pwd_salt))

	pwdChecked := dblayer.UserSignin(username,encpasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	token := GenToken(username)
	upRes := dblayer.UpdateToken(username,token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())

}

func GenToken(username string) string{
	//规则：md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x",time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username+ts+"_tokensalt"))
	return tokenPrefix + ts[:8]
}
