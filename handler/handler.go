package handler

import (
	cfg "FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"FileCloud/mq"
	"FileCloud/store/oss"
	"FileCloud/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	username := r.Form.Get("username")

	if r.Method == "GET" {
		//返回上传html页面
		data, err := ioutil.ReadFile("src/FileCloud/static/view/upload.html")
		if err != nil {
			io.WriteString(w, "internal ser ver error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流存储到本地目录
		//Form方法返回文件本体，文件头部信息，以及错误信息

		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}

		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "src/FileCloud/static/files/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		//创造一个新文件
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file,err:%s\n", err.Error())
			return
		}
		defer file.Close()

		//进行文件体替换
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file,err:%s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		//将文件写入到阿里云oss中
		// 读取文件流。
		//fd, err := os.Open(fileMeta.Location)
		//if err != nil {
		//	fmt.Println("Error:", err)
		//	os.Exit(-1)
		//}
		//开始同步写入
		/**
		以当前上传文件的用户名作为在oss存储上的文件夹名称
		即实行简单的一用户一文件夹策略
		*/
		newFile.Seek(0, 0)
		ossPath := username + "/" + fileMeta.FileName
		//err = oss.Bucket().PutObject(ossPath, fd)
		//if err != nil {
		//	fmt.Println(err.Error())
		//	w.Write([]byte(" Oss Upload failed"))
		//	return
		//}
		//fileMeta.Location = ossPath

		//通过RabbitMQ异步实现
		// 写入异步转移任务队列
		if !cfg.AsyncTransferEnable {
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				fmt.Println(err.Error())
				w.Write([]byte("Upload failed!"))
				return
			}
		} else {
			// 写入异步转移任务队列
			data := mq.TransferData{
				FileHash:     fileMeta.FileSha1,
				CurLocation:  fileMeta.Location,
				DestLocation: ossPath,
			}
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				cfg.TransExchangeName,
				cfg.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				// TODO: 当前发送转移信息失败，稍后重试
			} else {
				fmt.Println("成功发送消息：" + string(pubData))
			}
		}

		// 判断是否符合开启秒传功能
		fileMeta2, err := meta.GetFileMetaDB(fileMeta.FileSha1)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//更改文件保存地址为阿里云的云端存储地址
		fileMeta.Location = ossPath
		if fileMeta2.FileSha1 == "" {
			w.Write([]byte("Normal Upload"))
			// 保存信息到文件表
			_ = meta.UpdateFileMetaDB(fileMeta)
			// 保存信息到用户文件表
			dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		} else {
			w.Write([]byte("Fast Upload"))
			// 保存信息到用户文件表
			dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		}
	}
}

//上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

//根据hash值查询文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//批量查询对应用户的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 获取分页信息
	pageIndex, _ := strconv.Atoi(r.Form.Get("PageIndex"))
	pageSize, _ := strconv.Atoi(r.Form.Get("PageSize"))
	username := r.Form.Get("username")
	userFiles, err := dblayer.QueryUserFileMetas(username, pageIndex, pageSize)
	//给文件元信息体传递源文件名去前台
	for i := 0; i < len(userFiles); i++ {
		row, _ := dblayer.GetFileMeta(userFiles[i].FileHash)
		userFiles[i].RealName = row.FileName.String
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//文件下载
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	//解析url参数
	r.ParseForm()
	//得到需要的参数值
	fsha1 := r.Form.Get("filehash")
	//根据哈希值得到文件元信息
	fm := meta.GetFileMeta(fsha1)
	//根据文件元信息中的定位信息获取文件本体
	f, err := os.Open(fm.Location)
	fmt.Println(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//尝试打开文件本体，读取文件流，一般在浏览器中读取会直接提示下载
	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

//更新文件元信息接口（重命名）
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	//解析参数
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "2" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta, _ := meta.GetFileMetaDB(fileSha1)
	curFileMeta.FileName = newFileName
	dblayer.UpdateName(curFileMeta.FileName, curFileMeta.FileSha1)

	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(data)
}

//删除文件以及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//解析参数
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")
	//物理上的删除
	//TODO:存在有时失效问题
	fMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		fmt.Println(err.Error())
	}
	Location := "src/FileCloud/static/files/" + fMeta.FileName
	os.Remove(Location)

	//用户文件元信息的删除（删除用户文件表中的记录）
	dblayer.DeleteUserFile(fileSha1)

	//oss云上的删除
	bucket := oss.Bucket()
	err = bucket.DeleteObject(fMeta.Location)
	if err != nil {
		fmt.Println("Error:", err)
	}

	w.WriteHeader(http.StatusOK)

}

/**
文件秒传接口
*/
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

//生成文件下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	//从文件表中查找记录
	row, _ := dblayer.GetFileMeta(filehash)
	signedURL := oss.DownloadURL(row.FileAddr.String)
	w.Write([]byte(signedURL))
}

//获取数据库所有文件元数据信息
func GetAllFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// 获取分页信息
	pageIndex, _ := strconv.Atoi(r.Form.Get("PageIndex"))
	pageSize, _ := strconv.Atoi(r.Form.Get("PageSize"))
	fmt.Println(pageSize, pageIndex)

	userFiles, err := dblayer.GetAllFileMeta(pageIndex, pageSize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//给所有文件元信息体传递源文件名去前台
	for i := 0; i < len(userFiles); i++ {
		row, _ := dblayer.GetFileMeta(userFiles[i].FileHash)
		userFiles[i].RealName = row.FileName.String
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
