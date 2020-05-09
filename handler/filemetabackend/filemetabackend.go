package filemetabackend

import "net/http"

/**
文件元数据管理提供的功能有：
1、获取系统中所有文件元数据列表()
2、根据用户名查找文件元数据列表()
3、根据文件名查找文件元数据列表()
4、查看元数据变更记录
 */

// 可能可以通过阿里云来获取文件元数据
// 文件元信息（Object Meta）是对上传到OSS的文件的属性描述，分为两种：HTTP标准属性（HTTP Headers）和 User Meta（用户自定义元信息）。
// 文件元信息可以在各种方式上传时或者拷贝文件时进行设置。

// 从本地获取文件元数据信息
// 这里本地文件元信息为比较基础的文件信息为主
func GetAllFileMeta(w http.ResponseWriter, r *http.Request) {
}

// 从OSS上获取文件元数据信息
func GetFileMetaFormOSS(w http.ResponseWriter, r *http.Request) {
}

// 修改文件元信息
func UpdateFileMetaFormOSS(w http.ResponseWriter, r *http.Request) {
}

// 查看文件变更记录
func GetFileOperation(w http.ResponseWriter, r *http.Request) {
}