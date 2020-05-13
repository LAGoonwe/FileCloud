package main

import (
	"FileCloud/config"
	"FileCloud/handler"
	backendhandler "FileCloud/handler/filebackend"
	"FileCloud/handler/filemetabackend"
	"FileCloud/util"
	operationhandler "FileCloud/handler/operation"
	"fmt"
	"net/http"
)

func main() {
	// 静态资源处理
	// 设置静态资源目录
	//fsh := http.FileServer(http.Dir("src/FileCloud/static"))
	//http.Handle("/static/", http.StripPrefix("/static/", fsh))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticdir()))))

	// 搭建指定目录的本地服务器【指定目录即为config里的UploadPath】
	util.LocalFileServerWithUploadPath()

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

	// 重置密码相关
	http.HandleFunc("/user/checkauth", handler.CheckAuthHandler)
	http.HandleFunc("/user/resetpwd", handler.ResetUserPwd)
	//手机邮箱验证接口
	http.HandleFunc("/valide/email", handler.EmailValideHandler)
	http.HandleFunc("/valide/phone", handler.PhoneValideHandler)

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
	// 文件上传（用户简单上传）√
	http.HandleFunc("/filebackend/upload", handler.HTTPInterceptor(backendhandler.BackendUploadHandler))
	// 文件重命名√
	http.HandleFunc("/filebackend/update", handler.HTTPInterceptor(backendhandler.UpdateBackendUserFilesName))
	// 获取用户文件信息√
	http.HandleFunc("/filebackend/query", handler.HTTPInterceptor(backendhandler.QueryBackendUserFiles))
	// 文件下载（用户简单下载）√
	http.HandleFunc("/filebackend/download", handler.HTTPInterceptor(backendhandler.LocalDownLoadFile))
	// 文件删除√
	http.HandleFunc("/filebackend/delete", handler.HTTPInterceptor(backendhandler.DeleteFile))
	// 返回下载外链√
	http.HandleFunc("/filebackend/downloadurl", handler.HTTPInterceptor(backendhandler.GetDownLoadFileURL))
	// 秒传√
	http.HandleFunc("/filebackend/fastupload", handler.HTTPInterceptor(backendhandler.BackendTryFastUpload))
	// 修改文件内容
	http.HandleFunc("/filebackend/updatefilecontent", handler.HTTPInterceptor(backendhandler.UpdateFileContent))

	// 初始化分块√---
	http.HandleFunc("/filebackend/mpupload/init", handler.HTTPInterceptor(backendhandler.InitialMultipartUpload))
	// 上传文件分块√---
	http.HandleFunc("/filebackend/mpupload/uppart", handler.HTTPInterceptor(backendhandler.UploadPart))
	// 通知上传并合并√---
	http.HandleFunc("/filebackend/mpupload/complete", handler.HTTPInterceptor(backendhandler.CompleteUpload))
	// 通知取消上传√---
	http.HandleFunc("/filebackend/mpupload/canceluppart", handler.HTTPInterceptor(backendhandler.CancelUploadPart))
	// 查看分块上传状态√---
	http.HandleFunc("/filebackend/mpupload/uppartstatus", handler.HTTPInterceptor(backendhandler.MultipartUploadStatus))

	// 系统管理员通过名称模糊查询文件信息√---
	http.HandleFunc("/filebackend/querybyname", handler.HTTPInterceptor(backendhandler.GetBackendUserFilesByName))
	// 系统管理员阿里云范围下载√
	http.HandleFunc("/filebackend/downloadrange", handler.HTTPInterceptor(backendhandler.RangeDownLoadFile))
	// 系统管理员阿里云流式下载√
	http.HandleFunc("/filebackend/downloadlocal", handler.HTTPInterceptor(backendhandler.NormalDownLoadFile))
	// 系统管理员阿里云断点续传下载√
	http.HandleFunc("/filebackend/downloadpart", handler.HTTPInterceptor(backendhandler.PartDownLoadFile))

	// 系统管理员获取全部文件信息√
	http.HandleFunc("/filebackend/all", handler.HTTPInterceptor(backendhandler.GetAllBackendUserFiles))
	// 系统管理员上传字符√
	http.HandleFunc("/filebackend/uploadstring", handler.HTTPInterceptor(backendhandler.BackendUploadStringHandler))
	// 系统管理员上传文件流（同上传本地文件实现原理）√
	http.HandleFunc("/filebackend/uploadfile", handler.HTTPInterceptor(backendhandler.BackendUploadFileHandler))
	// 系统管理员追加上传×（有问题，ErrorCode=ObjectNotAppendable）
	http.HandleFunc("/filebackend/uploadappend", handler.HTTPInterceptor(backendhandler.BackendAppendUpload))
	// 系统管理员断点续传上传√
	http.HandleFunc("/filebackend/uploadpart", handler.HTTPInterceptor(backendhandler.BackendPartUploadHandler))
	// 系统管理员分片上传√
	http.HandleFunc("/filebackend/uploadcomplexpart", handler.HTTPInterceptor(backendhandler.ComplexBackendPartUploadHandler))
	// 系统管理员列举OSS的所有文件√
	http.HandleFunc("/filebackend/listallossfiles", handler.HTTPInterceptor(backendhandler.ListAllOSSFiles))
	// 系统管理员获取文件访问权限√
	http.HandleFunc("/filebackend/getossacl", handler.HTTPInterceptor(backendhandler.GetOSSFileACL))
	// 系统管理员设置文件访问权限√
	http.HandleFunc("/filebackend/setossacl", handler.HTTPInterceptor(backendhandler.SetOSSFileACL))
	// 系统管理员判断OSS文件是否存在√
	http.HandleFunc("/filebackend/isexistossfile", handler.HTTPInterceptor(backendhandler.IsExistOSSFile))
	// 系统管理员冻结文件或者取消冻结文件--
	http.HandleFunc("/filebackend/changefilestatus", handler.HTTPInterceptor(backendhandler.UpdateFileStatus))
	// 系统管理员更改所选文件状态√
	http.HandleFunc("/filebackend/updatefilestatus", handler.HTTPInterceptor(backendhandler.UpdateFileStatus))
	// 系统管理员复制所选文件√---
	http.HandleFunc("/filebackend/copyfile", handler.HTTPInterceptor(backendhandler.CopyFile))
	// 系统管理员移动所选文件√---
	http.HandleFunc("/filebackend/movefile", handler.HTTPInterceptor(backendhandler.MoveFile))
	// 系统管理员更改文件存储类型（暂时不实现）
	http.HandleFunc("/filebackend/changefilestore", handler.HTTPInterceptor(backendhandler.ChangeFileStore))

	// 文件元数据管理
	// 系统管理员获取所有文件元信息√
	http.HandleFunc("/filebackend/getallmeta", handler.HTTPInterceptor(filemetabackend.GetAllObjectMeta))
	// 系统管理员获取文件元信息√
	http.HandleFunc("/filebackend/getobjectmeta", handler.HTTPInterceptor(filemetabackend.GetObjectMeta))
	// 系统管理员修改文件元信息√
	http.HandleFunc("/filebackend/updateobjectmeta", handler.HTTPInterceptor(filemetabackend.UpdateObjectMeta))
	// 系统管理员查看文件元信息操作记录√
	http.HandleFunc("/filebackend/getobjectmetaoperation", handler.HTTPInterceptor(filemetabackend.GetObjectMetaOperation))

	// 操作记录查看
	// 系统管理员获取所有的操作记录√
	http.HandleFunc("/filebackend/getalloperations", handler.HTTPInterceptor(operationhandler.GetAllOperations))
	// 系统管理员获取指定文件的操作记录√
	http.HandleFunc("/filebackend/getoperationsbyfile", handler.HTTPInterceptor(operationhandler.GetOperationsByUserFileId))
	// 系统管理员获取某个用户的操作记录√
	http.HandleFunc("/filebackend/getoperationsbyuser", handler.HTTPInterceptor(operationhandler.GetOperationsByUserId))
	// 系统管理员获取某个时间段的操作记录√
	http.HandleFunc("/filebackend/getoperationsbytime", handler.HTTPInterceptor(operationhandler.GetOperationsByTime))
	// 系统管理员获取某个操作类型的操作记录√
	http.HandleFunc("/filebackend/getoperationsbytype", handler.HTTPInterceptor(operationhandler.GetOperationsByOperationType))
	// 系统管理员获取某个操作id对应的操作记录√
	http.HandleFunc("/filebackend/getoperationsbyid", handler.HTTPInterceptor(operationhandler.GetOperationsByOperationId))

	// Bucket管理（SDK已实现，考虑到安全性问题，暂时不实现）

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
