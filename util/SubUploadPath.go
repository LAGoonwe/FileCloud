package util

import (
	"FileCloud/config"
	"net/http"
	"strings"
)

/**
截取本地服务器路径，搭建上传路径的文件服务器，辅助本地文件预览
*/

func LocalFileServerWithUploadPath() {

	// 引入本地服务器路径并截取盘符，以：作为分隔
	// 取出盘符
	Disk := config.UploadPath[0:strings.IndexAny(config.UploadPath, ":")]
	// 取出路径
	FilePath := config.UploadPath[strings.IndexAny(config.UploadPath, ":")+1:]

	// 搭建本地文件服务器，文件预览使用
	http.Handle("/"+Disk+FilePath, http.StripPrefix("/"+Disk+FilePath, http.FileServer(http.Dir(config.UploadPath))))
}
