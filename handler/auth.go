package handler

import (
	cfg "FileCloud/config"
	"FileCloud/util"
	"fmt"
	"net/http"
)

//HTTPInterceptor: http请求拦截器
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")

			fmt.Println("{username:" + username + "\ttoken:" + token + "}")
			tokenVali, err := util.ParseToken(username, token, []byte(cfg.SecretKey))
			if err != nil {
				fmt.Println(err)
			}

			if len(username) <= 0 || !tokenVali {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h(w, r)
		})
}
