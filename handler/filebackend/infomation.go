package backendhandler

import (
	cfg "FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"FileCloud/store/oss"
	"FileCloud/util"
	"net/http"
	"os"
	"strconv"
)



// 批量查询对应用户的文件信息
// 这种大数据量的接口不用系统的全局内存变量
func QueryBackendUserFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	}

	r.ParseForm()

	username := r.Form.Get("username")
	pageIndex, _ := strconv.Atoi(r.Form.Get("PageIndex"))
	pageSize, _ := strconv.Atoi(r.Form.Get("PageSize"))

	backendUserFiles, err := dblayer.QueryBackendUserFiles(username, pageIndex, pageSize)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "获取文件信息失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 如果没数据的话，返回空值
	resp := util.RespMsg{
		Code: 1,
		Msg:  "获取文件信息成功！",
		Data: backendUserFiles,
	}
	w.Write(resp.JSONBytes())
}



// 重命名文件
// 重命名文件的操作跟移动文件相类似
// OSS本身没实现重命名的逻辑，如果需要实现，要使用拷贝对象的接口
func UpdateBackendUserFilesName(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filesha1 := r.Form.Get("filehash")
	// 这里原则上还需要做字符串校验
	newFileName := r.Form.Get("filename")

	// 判断参数是否合法
	params := make(map[string]string)
	params["username"] = username
	params["filename"] = newFileName
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

	// 判断数据库中是否已有该名称的文件
	result, err := dblayer.IsExistSameNameFile(username, filesha1)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "传入的参数不合法！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	if result {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "该文件名已存在！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// OSS移动文件
	// 首先需要查找原来的objectName，同时要限制住目的objectName的命名
	fileMeta, err := dblayer.GetFileByFileSha1(filesha1, username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "重命名文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	oldObjectName := fileMeta.FileRelLocation
	newObjectName := username + "/" + newFileName
	// oss先复制，后删除，这两步理应也是原子性操作
	// 这里假设能够完成
	_, err = oss.CopyFiles(oldObjectName, newObjectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "重命名文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	_, err = oss.DeleteOneFile(oldObjectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "重命名文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 本地服务器端移动文件
	oldLocalFileName := fileMeta.FileAbsLocation
	newLocalFileName := cfg.UploadPath + username + "/" + newFileName
	os.Rename(oldLocalFileName, newLocalFileName)

	// 更新数据库中的记录
	_, err = dblayer.UpdateFileName(filesha1, newFileName, newLocalFileName, newObjectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "重命名文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 信息更新会全局内存变量中
	oldFileMeta, err := CheckGlobalFileMeta(filesha1)
	oldFileMeta.FileName = newFileName
	oldFileMeta.FileRelLocation = newObjectName
	oldFileMeta.FileAbsLocation = newLocalFileName
	meta.UpdateBackendFileMeta(*oldFileMeta)

	resp := util.RespMsg{
		Code: 1,
		Msg:  "重命名成功！",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 管理员接口
// 获取系统中的所有用户的文件（还没增加拦截器进行身份校验）
// 这种大数据量的接口不用系统的全局内存变量
func GetAllBackendUserFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	limitp := r.Form.Get("limit")

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
	params["limit"] = limitp
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

	limit, _ := strconv.Atoi(limitp)
	allUserFiles, err := dblayer.GetAllUserFiles(limit)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "获取用户所有文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "获取用户所有文件时成功！",
		Data: allUserFiles,
	}
	w.Write(resp.JSONBytes())
}



// 根据用户名模糊检索文件（暂时没考虑传入多个用户名的情况）
// 这种大数据量的接口不用系统的全局内存变量
func GetBackendUserFilesByUserName(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	checkUsername := r.Form.Get("checkusername")
	limitp := r.Form.Get("limit")

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
	params["limit"] = limitp
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

	limit, _ := strconv.Atoi(limitp)
	matchFiles, err := dblayer.GetFilesByUserName(checkUsername, limit)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "根据用户名模糊查询文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 为空的话直接返回空
	resp := util.RespMsg{
		Code: 1,
		Msg:  "根据用户名模糊查询文件成功！",
		Data: matchFiles,
	}
	w.Write(resp.JSONBytes())
}

// 根据文件名模糊检索文件（暂时没考虑传入多个文件名的情况）
// 这种大数据量的接口不用系统的全局内存变量
func GetBackendUserFileByFileName(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filename := r.Form.Get("filename")
	limitp := r.Form.Get("limit")

	// 判断请求接口的用户是否是系统管理员
	_, err := CheckUserStatus(username)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "没有权限访问该接口",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 判断参数是否合法
	params := make(map[string]string)
	params["username"] = username
	params["filename"] = filename
	params["limit"] = limitp
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

	limit, _ := strconv.Atoi(limitp)
	matchFiles, err := dblayer.GetFilesByFileName(filename, limit)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "根据文件名模糊查询文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "根据文件名模糊查询文件成功！",
		Data: matchFiles,
	}
	w.Write(resp.JSONBytes())
}

// 返回所选文件的外链
func GetDownLoadFileURL(w http.ResponseWriter, r *http.Request) {
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
		resp := util.RespMsg{
			Code: -1,
			Msg:  "返回所选文件的外链失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	downloadURL, err := oss.GetDownLoadURL(fileMeta.FileRelLocation)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "返回所选文件的外链失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "返回所选文件的外链成功！",
		Data: downloadURL,
	}
	w.Write(resp.JSONBytes())
}

// 系统管理员列举OSS的所有文件
func ListAllOSSFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")

		// 判断请求接口的用户是否是系统管理员
		_, err := CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		allOSSFiles, err := oss.ListFiles()
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "列举所有OSS文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "列举所有OSS文件成功！",
			Data: allOSSFiles,
		}
		w.Write(resp.JSONBytes())
	}
}

// 系统管理员获取文件权限
func GetOSSFileACL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		filename := r.Form.Get("filename")

		// 判断请求接口的用户是否是系统管理员
		_, err := CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 判断参数是否合法
		params := make(map[string]string)
		params["username"] = username
		params["filename"] = filename
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

		objectName := username + "/" + filename
		acl, err := oss.GetFileACL(objectName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取文件权限失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取文件权限成功！",
			Data: acl,
		}
		w.Write(resp.JSONBytes())
	}
}

// 系统管理员设置文件权限
func SetOSSFileACL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		filename := r.Form.Get("filename")
		// 这里原则需要校验
		acl := r.Form.Get("acl")

		// 判断请求接口的用户是否是系统管理员
		_, err := CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 判断参数是否合法
		params := make(map[string]string)
		params["username"] = username
		params["filename"] = filename
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

		objectName := username + "/" + filename
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
			Msg:  "设置文件权限成功！",
			Data: acl,
		}
		w.Write(resp.JSONBytes())
	}
}



// 判断OSS文件是否存在
func IsExistOSSFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		filename := r.Form.Get("filename")

		// 判断请求接口的用户是否是系统管理员
		_, err := CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 判断参数是否合法
		params := make(map[string]string)
		params["username"] = username
		params["filename"] = filename
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

		objectName := username + "/" + filename
		result, err := oss.IsExistFile(objectName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "判断OSS文件是否存在失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "判断OSS文件是否存在成功！",
			Data: result,
		}
		w.Write(resp.JSONBytes())
	}
}
