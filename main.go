package main

import (
	"FileCloud/handler"
	"fmt"
	"net/http"
)


func main() {

	// 设置静态资源目录
	fsh := http.FileServer(http.Dir("src/FileCloud/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fsh))

	//文件操作路由设置
	http.HandleFunc("/file/upload",handler.UploadHandler)
	http.HandleFunc("/file/upload/suc",handler.UploadSucHandler)
	http.HandleFunc("/file/meta",handler.GetFileMetaHandler)
	http.HandleFunc("/file/download",handler.DownloadHandler)
	http.HandleFunc("/file/update",handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete",handler.FileDeleteHandler)

	//用户操作路由设置
	http.HandleFunc("/user/signup",handler.SignupHandler)
	http.HandleFunc("/user/signin",handler.SignInHandler)
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		fmt.Printf("Failed to start server,err:%s",err.Error())
	}
}
