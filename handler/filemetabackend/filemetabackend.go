package filemetabackend

import (
	common "FileCloud/common"
	dblayer "FileCloud/db"
	filebackend "FileCloud/handler/filebackend"
	"FileCloud/store/oss"
	"FileCloud/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

/**
文件元数据管理提供的功能有：
1、获取已有文件元信息列表()
2、获取特定文件元信息()
3、修改文件元信息()
4、查看元数据变更记录()
*/

// 可能可以通过阿里云来获取文件元数据
// 文件元信息（Object Meta）是对上传到OSS的文件的属性描述，分为两种：HTTP标准属性（HTTP Headers）和 User Meta（用户自定义元信息）。
// 文件元信息可以在各种方式上传时或者拷贝文件时进行设置。

// 获取已有文件元信息列表
func GetAllObjectMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	limitp := r.Form.Get("limit")
	limit, _ := strconv.Atoi(limitp)

	// 判断请求接口的用户是否是系统管理员
	_, err := filebackend.CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	allObjectMetas, err := dblayer.GetAllObjectMeta(limit)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "获取文件元信息列表失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "获取文件元信息列表成功！",
		Data: allObjectMetas,
	}
	w.Write(resp.JSONBytes())
}

// 从OSS获取文件元数据信息并存入本地
func GetObjectMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	objectName := r.Form.Get("objectname")

	// 判断请求接口的用户是否是系统管理员
	_, err := filebackend.CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 数据库查询ObjectName对应的文件对象是否存在
	backendUserFile, err := dblayer.GetFileByObjectName(objectName)
	fmt.Println(backendUserFile)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "从OSS获取元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	if backendUserFile.FileName == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "找不到该object！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	fmt.Println(backendUserFile)

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
	//userFileId := backendUserFile.Id
	// objectMeta类型为Http.Header，原始类型为map[string][]string
	meta, _ := json.Marshal(objectMeta)
	metaString := string(meta)
	userFileId := backendUserFile.Id
	metaId, err := dblayer.InsertFileMeta(userFileId, metaString)
	if err != nil {
		log.Println("元数据信息插入数据库失败！")
		resp := util.RespMsg{
			Code: -1,
			Msg:  "从OSS获取元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 异步：修改记录插入到数据库中
	detail := "用户：" + username + "，查看文件：" + backendUserFile.FileRelLocation + "的元信息"
	_, err = dblayer.InsertOperation(common.GetObjectMeta, metaId, backendUserFile.FileSha1, username, detail)
	if err != nil {
		// 不影响主要业务
		log.Println("操作记录插入到数据库失败！")
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "从OSS获取元信息成功！",
		Data: metaString,
	}
	w.Write(resp.JSONBytes())
}

// 修改文件元信息
// 自定义元信息，以X-Oss-Meta-为前缀的参数
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
	backendUserFile, err := dblayer.GetFileByObjectName(objectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "修改文件元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	if backendUserFile.FileName == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "找不到该object！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	fmt.Println(backendUserFile)

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
		log.Println("获取文件元信息失败！")
		resp := util.RespMsg{
			Code: -1,
			Msg:  "修改OSS文件元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	fmt.Println(objectMeta)
	meta, _ := json.Marshal(objectMeta)
	metaString := string(meta)
	userFileId := backendUserFile.Id
	metaId, err := dblayer.UpdateFileMeta(userFileId, metaString)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "修改OSS文件元信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 异步：修改记录插入到数据库中
	detail := "用户：" + username + "，修改文件：" + backendUserFile.FileRelLocation + "元信息"
	_, err = dblayer.InsertOperation(common.UpdateObjectMeta, metaId, backendUserFile.FileSha1, username, detail)
	if err != nil {
		// 不影响主要业务
		log.Println("记录插入到数据库失败！")
	}

	resp := util.RespMsg{
		Code: -1,
		Msg:  "修改OSS文件元信息成功！",
		Data: metaString,
	}
	w.Write(resp.JSONBytes())
}

// 查看文件元信息修改记录
func GetObjectMetaOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	metaId := r.Form.Get("metaid")
	metaIdInt, _ := strconv.Atoi(metaId)

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

	operations, err := dblayer.GetOperationByMetaId(metaIdInt)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "查看文件元信息修改记录失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: -1,
		Msg:  "查看文件元信息修改记录成功！",
		Data: operations,
	}
	w.Write(resp.JSONBytes())
}
