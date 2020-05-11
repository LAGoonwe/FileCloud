package backendhandler

import (
	cfg "FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"FileCloud/store/oss"
	"FileCloud/util"
	"log"
	"net/http"
	"os"
)

// 删除文件
// OSS同时也实现了删除多个文件的方法
// 除了删除本地服务器上的，还要删除OSS上的
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")

	// 判断参数是否合法
	params := make(map[string]string)
	params["username"] = username
	params["filehash"] = filehash
	_, err := CheckParams(params)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "传入的参数不合法！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	fileMeta, err := CheckGlobalFileMeta(filehash)
	if err != nil {
		log.Println(err.Error())
		resp := util.RespMsg{
			Code: -1,
			Msg:  "删除文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	objectName := fileMeta.FileRelLocation
	localFileName := fileMeta.FileAbsLocation
	// 这里本质要求是一个原子性操作，即要求本地服务器和OSS的文件都要被删除
	// 先进行OSS的删除，由于网络原因导致的错误概率比较高
	_, err = oss.DeleteOneFile(objectName)
	if err != nil {
		log.Println(err.Error())
		resp := util.RespMsg{
			Code: -1,
			Msg:  "删除文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// 删除本地服务器的文件
	os.Remove(localFileName)

	// 删除数据库中的记录
	_, err = dblayer.DeleteFile(filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "删除文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 删除全局对象中的内容
	meta.DeleteBackendFileMeta(filehash)

	resp := util.RespMsg{
		Code: 1,
		Msg:  "文件删除成功！",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 更改所选文件状态
func UpdateFileStatus(w http.ResponseWriter, r *http.Request) {
	// 冻结文件以后，用户无法对文件进行任何操作
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")

	// 判断请求接口的用户是否是系统管理员
	_, err := CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 判断参数是否合法
	params := make(map[string]string)
	params["username"] = username
	params["filehash"] = filehash
	_, err = CheckParams(params)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "传入的参数不合法！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	fileMeta, err := CheckGlobalFileMeta(filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "更改文件状态失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	newBackendUserFile, err := dblayer.UpdateFileStatus(fileMeta.FileSha1)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "更改文件状态失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	fileMeta.Status = newBackendUserFile.Status
	resp := util.RespMsg{
		Code: 1,
		Msg:  "success",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 移动所选文件
func MoveFile(w http.ResponseWriter, r *http.Request) {
	// 移动所选文件的位置
	// OSS目前不支持直接移动文件，需要把文件Copy到目的地，再删除源目的地的文件
	// 链接：https://help.aliyun.com/knowledge_detail/39622.html?spm=5176.11065259.1996646101.searchclickresult.21387cc2wKX0vs&aly_as=GMs3aQodB
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	// 这里是相对路径，这里原则上还要做字符串校验
	newFilePath := r.Form.Get("filepath")

	// 判断请求接口的用户是否是系统管理员
	_, err := CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 判断参数是否合法
	params := make(map[string]string)
	params["username"] = username
	params["filehash"] = filehash
	_, err = CheckParams(params)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "传入的参数不合法！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	fileMeta, err := CheckGlobalFileMeta(filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 先进行OSS端的文件移动
	oldObjectName := fileMeta.FileRelLocation
	destObjectName := username + "\\" + newFilePath

	// 本地服务器文件路径
	oldLocalFileName := fileMeta.FileAbsLocation
	destLocalFileName := cfg.UploadPath + username + "\\" + newFilePath
	_, err = oss.CopyFiles(oldObjectName, destObjectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 这个操作有风险，就是没有做原子性操作，理应是移动成功后加入到任务队列中，隔断时间再做删除
	_, err = oss.DeleteOneFile(oldObjectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	// 移动本地服务器文件，还没测试是否会删除源目的地的文件
	os.Rename(oldLocalFileName, destLocalFileName)

	// 把更新的数据刷新到全局变量中
	fileMeta.FileRelLocation = destObjectName
	fileMeta.FileAbsLocation = destLocalFileName

	resp := util.RespMsg{
		Code: 1,
		Msg:  "移动文件成功",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 修改文件内容（还未实现）
func UpdateFileContent(w http.ResponseWriter, r *http.Request) {
	// 这里的修改文件是：用户重新上传文件，覆盖相同位置的原文件
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}
}


// 设置文件访问权限
// 文件默认访问权限如果没设置的话，继承存储空间的访问权限
func UpdateFileACL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	// 这里原则上要做检验的，特别是这种权限设置的接口
	acl := r.Form.Get("acl")

	// 判断请求接口的用户是否是系统管理员
	_, err := CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 判断参数是否合法
	params := make(map[string]string)
	params["username"] = username
	params["filehash"] = filehash
	_, err = CheckParams(params)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "传入的参数不合法！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	fileMeta, err := CheckGlobalFileMeta(filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "设置文件权限失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	objectName := fileMeta.FileRelLocation

	_, err = oss.SetFileACL(objectName, acl)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "设置文件权限失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "success",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 更改文件存储类型
