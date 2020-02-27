package handler

import (
	"FileCloud/meta"
	"FileCloud/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request)  {

	if r.Method == "GET" {
		//返回上传html页面
		data, err := ioutil.ReadFile("src/FileCloud/static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	}else if r.Method == "POST" {
		//接收文件流存储到本地目录
		//Form方法返回文件本体，文件头部信息，以及错误信息
		file,head,err := r.FormFile("file")
		if err != nil{
			fmt.Printf("Failed to get data,err:%s\n",err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "src/FileCloud/static/files/"+head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		//创造一个新文件
		newFile,err := os.Create(fileMeta.Location)
		if err != nil{
			fmt.Printf("Failed to create file,err:%s\n",err.Error())
			return
		}
		defer file.Close()

		//进行文件体替换
		fileMeta.FileSize,err = io.Copy(newFile,file)
		if err != nil{
			fmt.Printf("Failed to save data into file,err:%s\n",err.Error())
			return
		}

		newFile.Seek(0,0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMetaDB(fileMeta)

		http.Redirect(w,r,"/file/upload/suc",http.StatusFound)
	}
}

//上传已完成
func UploadSucHandler(w http.ResponseWriter,r *http.Request)  {
	io.WriteString(w,"Upload finished!")
}

//根据hash值查询文件元信息
func GetFileMetaHandler(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	fMeta,err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data,err := json.Marshal(fMeta)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter,r *http.Request)  {
	//解析url参数
	r.ParseForm()
	//得到需要的参数值
	fsha1 := r.Form.Get("filehash")
	//根据哈希值得到文件元信息
	fm := meta.GetFileMeta(fsha1)
	//根据文件元信息中的定位信息获取文件本体
	f,err := os.Open(fm.Location)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//尝试打开文件本体，读取文件流，一般在浏览器中读取会直接提示下载
	data,err :=ioutil.ReadAll(f)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

//更新文件元信息接口（重命名）
func FileMetaUpdateHandler(w http.ResponseWriter,r *http.Request)  {
	//解析参数
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0"{
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST"{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName=newFileName
	meta.UpdateFileMeta(curFileMeta)

	w.WriteHeader(http.StatusOK)
	data,err :=json.Marshal(curFileMeta)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(data)
}


//删除文件以及元信息
func FileDeleteHandler(w http.ResponseWriter,r *http.Request)  {
	//解析参数
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")

	//物理上的删除
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)

	//元信息，即索引的删除
	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)
	
}

