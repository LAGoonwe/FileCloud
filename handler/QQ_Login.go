package handler

import (
	"FileCloud/config"
	"FileCloud/db"
	"FileCloud/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type PrivateInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"open_id"`
}

//Get Authorization Code
func GetAuthCode(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", config.AppId)
	str := fmt.Sprintf("%s&redirect_uri=%s", params.Encode(), config.RedirectURI)
	loginURL := fmt.Sprintf("%s?%s", "https://graph.qq.com/oauth2.0/authorize", str)

	http.Redirect(w, r, loginURL, http.StatusFound)
}

// 2. Get Access Token
func GetToken(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", config.AppId)
	params.Add("client_secret", config.AppKey)
	params.Add("code", code)
	str := fmt.Sprintf("%s&redirect_uri=%s", params.Encode(), config.RedirectURI)
	loginURL := fmt.Sprintf("%s?%s", "https://graph.qq.com/oauth2.0/token", str)

	response, err := http.Get(loginURL)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	defer response.Body.Close()

	bs, _ := ioutil.ReadAll(response.Body)
	body := string(bs)

	resultMap := convertToMap(body)

	info := &PrivateInfo{}
	info.AccessToken = resultMap["access_token"]
	info.RefreshToken = resultMap["refresh_token"]
	info.ExpiresIn = resultMap["expires_in"]

	GetOpenId(info, w, r)
}

// 3. Get OpenId
func GetOpenId(info *PrivateInfo, w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s", "https://graph.qq.com/oauth2.0/me", info.AccessToken))
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	defer resp.Body.Close()

	bs, _ := ioutil.ReadAll(resp.Body)
	body := string(bs)
	info.OpenId = body[45:77]

	GetUserInfo(info, w, r)
}

// 4. Get User info
func GetUserInfo(info *PrivateInfo, w http.ResponseWriter, r *http.Request) {

	params := url.Values{}
	params.Add("access_token", info.AccessToken)
	params.Add("openid", info.OpenId)
	params.Add("oauth_consumer_key", config.AppId)

	uri := fmt.Sprintf("https://graph.qq.com/user/get_user_info?%s", params.Encode())
	resp, err := http.Get(uri)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	defer resp.Body.Close()

	bs, _ := ioutil.ReadAll(resp.Body)
	classDetailMap := make(map[string]string)
	_ = json.Unmarshal(bs, &classDetailMap)
	w.Write([]byte(classDetailMap["nickname"]))

	//查询用户表中是否已经有该用户名存在，如果不存在则代表是首次扫码登录，新增一条记录
	user, err := db.GetUserInfo(classDetailMap["nickname"])
	if err != nil {
		fmt.Println(err.Error())
	}
	if user.Username == "" {
		//插入新数据
		//对用户密码进行哈希的加密处理
		//enc_passwd := util.Sha1([]byte(classDetailMap["nickname"] + pwd_salt))
		//_ = db.UserSignup(classDetailMap["nickname"], enc_passwd)
	} else {
		//如果用户表存在该用户名，则代表之前扫码登陆过系统，取出相应的账号密码即可
		fmt.Println(user.Username)
	}

	//token := GenToken(classDetailMap["nickname"])
	//upRes := db.UpdateToken(classDetailMap["nickname"], token)
	//if !upRes {
	//	w.Write([]byte("FAILED"))
	//	return
	//}

	resp2 := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Status   int
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: classDetailMap["nickname"],
			Status:   user.Status,
			//Token:    token,
		},
	}
	w.Write(resp2.JSONBytes())
}

func convertToMap(str string) map[string]string {
	var resultMap = make(map[string]string)
	values := strings.Split(str, "&")
	for _, value := range values {
		vs := strings.Split(value, "=")
		resultMap[vs[0]] = vs[1]
	}
	return resultMap
}
