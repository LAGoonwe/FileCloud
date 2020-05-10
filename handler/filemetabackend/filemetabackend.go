package filemetabackend

import (
	filebackend "FileCloud/handler/filebackend"
	dblayer "FileCloud/db"
	oss "FileCloud/store/oss"
	"FileCloud/util"
	"fmt"
	"net/http"
)

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

// 从OSS获取文件元数据信息并存入本地
func GetObjectMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	objectName := r.Form.Get("objectName")

	// 判断请求接口的用户是否是系统管理员
	_, err := filebackend.CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 数据库查询ObjectName是否存在
	result, err := dblayer.IsExistObjectName(username, objectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "从OSS获取元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	if !result {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "找不到该object！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 查询OSS中该Object的元信息，并存入数据库中
	objectMeta, err := oss.GetFileMeta(objectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "从OSS获取元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	fmt.Println(objectMeta)

	// 数据库以一个字段来存储元信息

	resp := util.RespMsg{
		Code: 1,
		Msg:  "从OSS获取元信息成功！",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 修改文件元信息
func UpdateObjectMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	// TODO 这里暂时还没跟前端对齐，因此假定仅传入一对值
	r.ParseForm()
	username := r.Form.Get("username")
	objectName := r.Form.Get("objectname")
	objectKey := r.Form.Get("metakey")
	objectValue := r.Form.Get("metavalue")

	// 判断请求接口的用户是否是系统管理员
	_, err := filebackend.CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 数据库查询ObjectName是否存在
	result, err := dblayer.IsExistObjectName(username, objectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "从OSS获取元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	if !result {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "找不到该object！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 修改的元信息
	var props map[string]string
	props = make(map[string]string, 0)
	props[objectKey] = objectValue
	// 向OSS发送修改元信息请求
	_, err = oss.ModifyFileMeta(objectName, props)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "修改OSS文件元信息失败",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 向OSS获取新的文件元数据，并存入数据库中
	objectMeta, err := oss.GetFileMeta(objectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "修改OSS文件元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	fmt.Println(objectMeta)

	// 存入数据库中

	resp := util.RespMsg{
		Code: -1,
		Msg:  "修改OSS文件元信息成功！",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 查看文件元信息修改记录
func GetObjectMetaRecord(w http.ResponseWriter, r *http.Request) {

}
