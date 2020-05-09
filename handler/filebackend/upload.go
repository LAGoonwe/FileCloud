package backendhandler

import (
	cfg "FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"FileCloud/store/oss"
	"FileCloud/util"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// 本地上传文件
func BackendUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")

		file, head, err := r.FormFile("file")
		if err != nil {
			log.Println("UploadHandler Failed")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		defer file.Close()

		// 到数据库中查找该对象是否唯一
		// 不允许同一用户上传文件名或hash值或文件名和hash值都相同的文件
		filename := head.Filename
		// 这里先判断数据库中是否存在同名文件即可
		result, err := dblayer.IsExistSameNameFile(username, filename)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		// 存在同名文件
		if result {
			log.Println("重复上传同名文件！")
			resp := util.RespMsg{
				Code: -1,
				Msg:  "不可重复上传文件名相同的文件！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 创建一个本地服务器新文件
		absoluteFileLocation := cfg.UploadPath + filename
		newFile, err := os.Create(absoluteFileLocation)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		filesize, err := io.Copy(newFile, file)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		newFile.Seek(0, 0)
		filesha1 := util.FileSha1(newFile)

		newFile.Seek(0, 0)
		// 判断数据库中有没有相同内容的文件
		result, err = dblayer.IsExistSameContentFile(username, filesha1)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 存在内容相同的文件
		if result {
			log.Println("重复上传同内容文件！")
			resp := util.RespMsg{
				Code: -1,
				Msg:  "不可重复上传内容相同的文件！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		objectName := username + "/" + filename
		backendFile := meta.BackendFile{
			UserName:        username,
			FileSha1:        filesha1,
			FileName:        filename,
			FileSize:        filesize,
			FileAbsLocation: absoluteFileLocation,
			FileRelLocation: objectName,
			Status:          1,
		}

		// TODO 这块还需要完善 通过RabbitMQ实现异步上传
		if !cfg.AsyncTransferEnable {
			// OSS不需要判断上传的文件是否已存在OSS，因为数据库端和OSS保持一致，前面已做判断
			err = oss.Bucket().PutObject(objectName, newFile)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "上传文件失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}
		} else {
			// 写入异步转移任务队列
		}

		// 同步到全局内存变量
		meta.BackendFiles[filesha1] = backendFile
		fmt.Println(meta.BackendFiles[filesha1])
		// 上传完毕插入数据库
		_, err = dblayer.OnBackendUserFileUploadFinished(
			backendFile.UserName, backendFile.FileSha1, backendFile.FileName, backendFile.FileAbsLocation, backendFile.FileRelLocation, backendFile.FileSize, backendFile.Status)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
	}
}

// 阿里云上传字符串


// 阿里云上传文件流



// 阿里云上传本地文件



// 阿里云追加上传



// 阿里云断点续传上传



// 阿里云分片上传