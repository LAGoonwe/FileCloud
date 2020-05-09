package main

import (
	"FileCloud/config"
	"FileCloud/handler"
	backendhandler "FileCloud/handler/filebackend"
	"fmt"
	"net/http"
)

func main() {
	// 静态资源处理
	// 设置静态资源目录
	//fsh := http.FileServer(http.Dir("src/FileCloud/static"))
	//http.Handle("/static/", http.StripPrefix("/static/", fsh))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticdir()))))

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

	//手机邮箱验证接口
	http.HandleFunc("/valide/email", handler.EmailValideHandler)

	//第三方控制器
	http.HandleFunc("/toLogin", handler.GetAuthCode)
	http.HandleFunc("/qqLogin", handler.GetToken)

	/**
	= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =
	文件、文件元数据模块新接口
	/filebackend 为文件模块接口
	/filemetabackend 为文件元数据模块接口
	= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =
	*/
	// 文件上传（用户简单上传）
	http.HandleFunc("/filebackend/upload", handler.HTTPInterceptor(backendhandler.BackendUploadHandler))
	// 文件重命名
	http.HandleFunc("/filebackend/update", handler.HTTPInterceptor(backendhandler.UpdateBackendUserFilesName))
	// 获取用户文件信息
	http.HandleFunc("/filebackend/query", handler.HTTPInterceptor(backendhandler.QueryBackendUserFiles))
	// 文件下载（用户简单下载）
	http.HandleFunc("/filebackend/download", handler.HTTPInterceptor(backendhandler.LocalDownLoadFile))
	// 文件删除
	http.HandleFunc("/filebackend/delete", handler.HTTPInterceptor(backendhandler.DeleteFile))
	// 返回下载外链
	http.HandleFunc("/filebackend/downloadurl", handler.HTTPInterceptor(backendhandler.GetDownLoadFileURL))
	//// 秒传
	//http.HandleFunc("/filebackend/fastupload", handler.HTTPInterceptor)
	//// 分块上传
	//http.HandleFunc("/filebackend/mpupload/init", handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	//http.HandleFunc("/filebackend/mpupload/uppart", handler.HTTPInterceptor(handler.UploadPartHandler))
	//http.HandleFunc("/filebackend/mpupload/complete", handler.HTTPInterceptor(handler.CompleteUploadHandler))

	// 系统管理员通过用户名模糊查询文件信息
	http.HandleFunc("/filebackend/querybyusername", handler.HTTPInterceptor(backendhandler.GetBackendUserFilesByUserName))
	// 系统管理员通过文件名模糊查询文件信息
	http.HandleFunc("/filebackend/querybyfilename", handler.HTTPInterceptor(backendhandler.GetBackendUserFileByFileName))
	// 系统管理员阿里云范围下载
	http.HandleFunc("/filebackend/downloadrange", handler.HTTPInterceptor(backendhandler.RangeDownLoadFile))
	// 系统管理员阿里云本地下载
	http.HandleFunc("/filebackend/downloadlocal", handler.HTTPInterceptor(backendhandler.NormalDownLoadFile))
	// 系统管理员阿里云断点续传下载
	http.HandleFunc("/filebackend/downloadpart", handler.HTTPInterceptor(backendhandler.PartDownLoadFile))
	// 系统管理员获取全部文件信息
	http.HandleFunc("/filebackend/all", handler.HTTPInterceptor(backendhandler.GetAllBackendUserFiles))
	// 系统管理员上传字符串
	// 系统管理员上传文件流
	// 系统管理员上传本地文件
	// 系统管理员追加上传
	// 系统管理员断点续传上传
	// 系统管理员分片上传

	// 文件元数据管理

	// Bucket管理（暂时不实现）

	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", config.UploadServiceHost)

	err := http.ListenAndServe(config.UploadServiceHost, nil)
	if err != nil {
		fmt.Println("Failed to start server, err: %s", err.Error())
	}
}

func staticdir() string {
	dir := config.StaticPath
	return dir
}
