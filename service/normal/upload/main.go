package main

import (
	"FileCloud/config"
	"FileCloud/handler"
	"fmt"
	"net/http"
)

func main() {
	// 静态资源处理
	// 设置静态资源目录
	fsh := http.FileServer(http.Dir("src/FileCloud/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fsh))

	// 文件相关
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/upload/suc", handler.HTTPInterceptor(handler.UploadSucHandler))
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/all", handler.HTTPInterceptor(handler.GetAllFileMetaHandler))
	// 秒传接口
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))

	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(handler.DownloadURLHandler))

	// 分块上传
	http.HandleFunc("/file/mpupload/init", handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete", handler.HTTPInterceptor(handler.CompleteUploadHandler))

	// 用户相关
	http.HandleFunc("/", handler.SignInHandler)
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))
	http.HandleFunc("/user/update", handler.HTTPInterceptor(handler.UpdateUserInfo))
	http.HandleFunc("/status/update", handler.HTTPInterceptor(handler.UpdateUserStatus))
	http.HandleFunc("/user/query", handler.HTTPInterceptor(handler.UserQueryHandler))
	http.HandleFunc("/user/addAdmin", handler.AddAdmin)
	http.HandleFunc("/user/delete", handler.HTTPInterceptor(handler.DeleteUserHandler))

	//第三方控制器
	http.HandleFunc("/toLogin", handler.GetAuthCode)
	http.HandleFunc("/qqLogin", handler.GetToken)

	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", config.UploadServiceHost)

	err := http.ListenAndServe(config.UploadServiceHost, nil)
	if err != nil {
		fmt.Println("Failed to start server, err: %s", err.Error())
	}
}
