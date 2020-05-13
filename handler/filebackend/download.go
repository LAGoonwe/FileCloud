package backendhandler

import (
	cfg "FileCloud/config"
	"FileCloud/store/oss"
	"FileCloud/util"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

// 流式下载所选文件(提供多种下载方式，同时提供多选下载)
// 通过阿里云下载可能比通过本地服务器传输要快
func NormalDownLoadFile(w http.ResponseWriter, r *http.Request) {
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
			Msg:  "流式下载出现问题！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	objectname := fileMeta.FileRelLocation
	filename := fileMeta.FileName
	data, err := oss.DownLoadStream(objectname)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "流式下载出现问题！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "流式下载出现问题！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\"" + filename + "\"")
	// 返回文件字节大小，辅助前端下载进度条实现
	w.Header().Set("Content-Length", strconv.FormatInt(fileMeta.FileSize, 10))
	w.Write([]byte(data))
}

// 阿里云范围下载
func RangeDownLoadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	startInt, _ := strconv.ParseInt(start, 10, 64)
	endInt, _ := strconv.ParseInt(end, 10, 64)

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
			Msg:  "范围下载文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}

	objectName := fileMeta.FileRelLocation
	fileName := fileMeta.FileName
	data, err := oss.DownLoadRangeFile(objectName, startInt, endInt)

	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "范围下载文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\"" + fileName + "\"")
	w.Write([]byte(data))
}

// 阿里云断点续传下载
// 下载到服务器本地文件
// 服务器在打开本地文件传输流
func PartDownLoadFile(w http.ResponseWriter, r *http.Request) {
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
			Msg:  "断点续传下载失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	objectName := fileMeta.FileRelLocation
	fileName := fileMeta.FileName
	preFilePath := cfg.PreUploadPath + username + "\\" + fileName
	fmt.Println(preFilePath)
	_, err = oss.DownLoadPartsFile(objectName, preFilePath, 3)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "断点续传下载失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	preLocalFile, err := os.Open(preFilePath)
	if err != nil {
		log.Println("PartDownLoadFile Error")
		resp := util.RespMsg{
			Code: -1,
			Msg:  "断点续传下载失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	defer preLocalFile.Close()

	data, err := ioutil.ReadAll(preLocalFile)
	fmt.Println(data)
	// 读取完文件后删除临时文件
	if err != nil {
		// 报错的时候删除临时文件
		err = os.Remove(preFilePath)
		if err != nil {
			log.Println("OS PreLocalFile Remove Error")
			log.Println(err.Error())
		}
		log.Println("DownLocalFile ReadAll Error")
		resp := util.RespMsg{
			Code: -1,
			Msg:  "断点续传下载失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 临时文件的删除出现错误，文件会被另外的进程占用
	//err = os.Remove(preFilePath)
	//if err != nil {
	//	log.Println("OS PreLocalFile Remove Error")
	//	log.Println(err.Error())
	//}

	// TODO 前端需要解决如何处理下载方式（当前前端是点击下载后直接触发下载到浏览器指定路径）
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("content-disposition", "attachment; filename=\"" + fileName + "\"")
	// 返回文件字节大小，辅助前端下载进度条实现
	w.Header().Set("Content-Length", strconv.FormatInt(fileMeta.FileSize, 10))
	w.Write(data)
}

// 本地下载方式
// 通用接口
func LocalDownLoadFile(w http.ResponseWriter, r *http.Request) {
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

	// 用户首次下载本地文件时, 将要下载的对象放入系统初始化的文件信息对象中
	fileMeta, err := CheckGlobalFileMeta(filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "简单下载文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	localFilePath := fileMeta.FileAbsLocation
	fileName := fileMeta.FileName
	localFile, err := os.Open(localFilePath)
	if err != nil {
		log.Println("DownLocalFile Open Error")
		resp := util.RespMsg{
			Code: -1,
			Msg:  "简单下载文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
	defer localFile.Close()

	// 前端直接提示下载
	data, err := ioutil.ReadAll(localFile)
	if err != nil {
		log.Println("DownLocalFile ReadAll Error")
		resp := util.RespMsg{
			Code: -1,
			Msg:  "简单下载文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fileName+"\"")
	// 返回文件字节大小，辅助前端下载进度条实现
	w.Header().Set("Content-Length", strconv.FormatInt(fileMeta.FileSize, 10))
	w.Write(data)
}

// 压缩下载
// 暂时没必要实现，性能不是特别好，而且麻烦
