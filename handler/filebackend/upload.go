package backendhandler

import (
	cfg "FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"FileCloud/store/oss"
	"FileCloud/util"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)



// 本地上传文件
// 这里也是可以实现秒传的
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
		absoluteFileLocation := cfg.UploadPath + username + "/" + filename
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
		// 上传完毕插入数据库，这里要确保不会发生异常
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



// 秒传
// 用户提供别人的文件hash值
func BackendTryFastUpload(w http.ResponseWriter, r *http.Request) {
	// 文件上传接口其实也可以走秒传的逻辑
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
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

		// 先判断用户是否已上传内容相同的文件
		result, err := dblayer.IsExistSameContentFile(username, filehash)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "秒传失败！",
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

		// 查找数据库是否已有该hash对应的文件
		fileMeta, err := dblayer.GetLocalFile(filehash)
		if fileMeta.UserName == "" {
			// 没有找到对应的hash文件，只能走普通上传文件方式，这里简单实现，直接让用户走普通上传方法
			log.Println("找不到已有文件记录！")
			resp := util.RespMsg{
				Code: -1,
				Msg:  "找不到已有文件记录，请走普通文件上传通道！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 不允许同一用户上传文件名或hash值或文件名和hash值都相同的文件
		// 判断数据库中是否存在同名文件即可
		result, err = dblayer.IsExistSameNameFile(username, fileMeta.FileName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "秒传上传文件失败！",
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

		// 创建一个文件对象
		filename := fileMeta.FileName
		fileSize := fileMeta.FileSize
		objectName := username + "/" + filename
		localFileName := cfg.UploadPath + username + "/" + filename

		backendFile := meta.BackendFile{
			UserName:        username,
			FileSha1:        filehash,
			FileName:        filename,
			FileSize:        fileSize,
			FileAbsLocation: localFileName,
			FileRelLocation: objectName,
			Status:          1,
		}

		// 这里如果找到了秒传文件记录的话，应该强制走异步加快上传速度的
		if !cfg.AsyncTransferEnable {
			// OSS不需要判断上传的文件是否已存在OSS，因为数据库端和OSS保持一致，前面已做判断
			file, err := os.Open(fileMeta.FileAbsLocation)
			if err != nil {
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "秒传失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}

			// 创建本地文件
			newFile, err := os.Create(localFileName)
			if err != nil {
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "秒传失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}

			_, err = io.Copy(newFile, file)
			if err != nil {
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "秒传失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}

			newFile.Seek(0, 0)
			// 将文件上传到OSS
			err = oss.Bucket().PutObject(objectName, newFile)
			if err != nil {
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "秒传失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}

			// TODO 这里存在问题还是，对于秒传来说改变了文件的路径，其实不应该包含用户名
			meta.BackendFiles[filehash] = backendFile
			fmt.Println(meta.BackendFiles[filehash])
			_, err = dblayer.OnBackendUserFileUploadFinished(
				backendFile.UserName, backendFile.FileSha1, backendFile.FileName, backendFile.FileAbsLocation, backendFile.FileRelLocation, backendFile.FileSize, backendFile.Status)
			if err != nil {
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "秒传失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}

			resp := util.RespMsg{
				Code: -1,
				Msg:  "秒传成功！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
		} else {
			// 写入异步转移任务队列
			resp := util.RespMsg{
				Code: -1,
				Msg:  "该功能还在调试！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
	}
}



// 阿里云上传字符串（上传Byte数组同理实现）
func BackendUploadStringHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		username := r.Form.Get("username")
		// 用户提供文件名
		filename := r.Form.Get("filename")
		// 文件中的字符串内容
		content := r.Form.Get("content")

		params := make(map[string]string)
		params["username"] = username
		params["filename"] = filename
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

		// 判断数据库中是否存在同名文件即可
		result, err := dblayer.IsExistSameNameFile(username, filename)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云字符串文件失败！",
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

		objectName := username + "/" + filename
		localFileName := cfg.UploadPath + username + "/" + filename
		preLocalFileName := cfg.PreUploadPath + username + "/" + filename
		// 创建本地服务器文件，并向本地服务器文件写入字符串内容
		newFile, err := os.Create(localFileName)
		preNewFile, err := os.Create(preLocalFileName)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云字符串文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		err = ioutil.WriteFile(preLocalFileName, []byte(content), 0777)
		filesize, err := io.Copy(newFile, preNewFile)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云字符串文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		filesha1 := util.FileSha1(newFile)
		newFile.Seek(0, 0)
		// 判断数据库中有没有相同内容的文件
		result, err = dblayer.IsExistSameContentFile(username, filesha1)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云字符串文件失败！",
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

		// 删除掉刚才的备用文件
		err = os.Remove(preLocalFileName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云字符串文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		backendFile := meta.BackendFile{
			UserName:        username,
			FileSha1:        filesha1,
			FileName:        filename,
			FileSize:        filesize,
			FileAbsLocation: localFileName,
			FileRelLocation: objectName,
			Status:          1,
		}

		// TODO 这块还需要完善 通过RabbitMQ实现异步上传
		if !cfg.AsyncTransferEnable {
			// OSS不需要判断上传的文件是否已存在OSS，因为数据库端和OSS保持一致，前面已做判断
			_, err = oss.UploadString(objectName, content)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "上传阿里云字符串文件失败！",
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
		// 上传完毕插入数据库，这里要确保不会发生异常
		_, err = dblayer.OnBackendUserFileUploadFinished(
			backendFile.UserName, backendFile.FileSha1, backendFile.FileName, backendFile.FileAbsLocation, backendFile.FileRelLocation, backendFile.FileSize, backendFile.Status)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云字符串文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
	}
}



// 阿里云上传文件流
func BackendUploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		file, head, err := r.FormFile("file")
		if err != nil {
			log.Println("BackendUploadFileHandler Failed")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云文件流失败！",
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
		absoluteFileLocation := cfg.UploadPath + username + "/" + filename
		newFile, err := os.Create(absoluteFileLocation)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传阿里云文件流失败！",
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
				Msg:  "上传阿里云文件流失败！",
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
			_, err = oss.UploadFileSteam(objectName, absoluteFileLocation)
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
		// 上传完毕插入数据库，这里要确保不会发生异常
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
	}
}



// 阿里云上传本地文件（同上传文件流一样）



// 阿里云追加末尾上传
// TODO 这里业务实现的话，考虑数据库表加字段
func BackendAppendUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		filename := r.Form.Get("filename")
		// 现在默认只支持内容追加到末尾
		// 追加到指定位置的操作还需要自行实现，同时对性能影响较大
		//startPos := r.Form.Get("start")
		appendValue := r.Form.Get("append")

		// 判断参数是否合法
		params := make(map[string]string)
		params["username"] = username
		params["filehash"] = filename
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
		// 传入的内容不能为空
		if appendValue == "" {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "传入的参数不合法！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 判断该文件是否已存在
		result, err := dblayer.IsExistSameNameFile(username, filename)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云追加上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		// 不存在文件
		if !result {
			log.Println("追加文件不存在")
			resp := util.RespMsg{
				Code: -1,
				Msg:  "追加内容的文件不存在！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		result, fileMeta, err := dblayer.GetFileByUserNameAndFileName(username, filename)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		objectName := fileMeta.FileRelLocation
		localFileName := fileMeta.FileAbsLocation

		// 先对本地文件做追加处理
		file, err := os.OpenFile(localFileName, os.O_WRONLY, 0644)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云追加上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		// 查找文件末尾的偏移量
		end, _ := file.Seek(0, os.SEEK_END)
		// 从末尾开始写入内容
		_, err = file.WriteAt([]byte(appendValue), end)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云追加上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 重新计算filehash和filesize
		preLocalFileName := cfg.PreUploadPath + username + "/" + filename
		preNewFile, err := os.Create(preLocalFileName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云追加上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		filesize, err := io.Copy(preNewFile, file)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云追加上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		filesha1 := util.FileSha1(preNewFile)
		// 判断数据库中有没有相同内容的文件
		result, err = dblayer.IsExistSameContentFile(username, filesha1)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "不可重复上传内容相同的文件！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		os.Remove(preLocalFileName)

		// 找出原来的对象
		oldFileMeta, err := CheckGlobalFileMeta(filesha1)
		oldFileMeta.FileSize = filesize
		oldFileMeta.FileSha1 = filesha1

		// TODO 这块还需要完善 通过RabbitMQ实现异步上传
		if !cfg.AsyncTransferEnable {
			// OSS不需要判断上传的文件是否已存在OSS，因为数据库端和OSS保持一致，前面已做判断
			// 返回的下一个参数为下次追加内容的位置
			_, _, err := oss.AppendFile(objectName, appendValue, end)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "阿里云追加上传文件失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}
		} else {
			// 写入异步转移任务队列
		}

		// 同步到全局内存变量
		meta.UpdateBackendFileMeta(*oldFileMeta)
		// 需要更新数据库记录
		_, err = dblayer.UpdateFileSizeAndFileHash(username, filename, filesha1, filesize)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云追加上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
	}
}



// 阿里云断点续传上传
func BackendPartUploadHandler(w http.ResponseWriter, r *http.Request) {
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
				Msg:  "阿里云断点续传上传文件失败！",
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
				Msg:  "阿里云断点续传上传文件失败",
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
		absoluteFileLocation := cfg.UploadPath + username + "/" + filename
		newFile, err := os.Create(absoluteFileLocation)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云断点续传上传文件失败！",
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
				Msg:  "阿里云断点续传上传文件失败！",
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
				Msg:  "阿里云断点续传上传文件失败！",
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
			_, err = oss.PartUpload(objectName, absoluteFileLocation, 3)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "阿里云断点续传上传文件失败！",
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
		// 上传完毕插入数据库，这里要确保不会发生异常
		_, err = dblayer.OnBackendUserFileUploadFinished(
			backendFile.UserName, backendFile.FileSha1, backendFile.FileName, backendFile.FileAbsLocation, backendFile.FileRelLocation, backendFile.FileSize, backendFile.Status)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云断点续传上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: -1,
			Msg:  "阿里云断点续传上传文件成功！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}
}



// 阿里云分片上传
func ComplexBackendPartUploadHandler(w http.ResponseWriter, r *http.Request) {
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
				Msg:  "阿里云分片上传文件失败！",
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
				Msg:  "阿里云分片上传文件失败！",
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
		absoluteFileLocation := cfg.UploadPath + username + "/" + filename
		newFile, err := os.Create(absoluteFileLocation)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云分片上传文件失败！",
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
				Msg:  "阿里云分片上传文件失败！",
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
				Msg:  "阿里云分片上传文件失败！",
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
			_, err = oss.ComplexPartUpload(objectName, absoluteFileLocation, 3)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "阿里云分片上传文件失败！",
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
		// 上传完毕插入数据库，这里要确保不会发生异常
		_, err = dblayer.OnBackendUserFileUploadFinished(
			backendFile.UserName, backendFile.FileSha1, backendFile.FileName, backendFile.FileAbsLocation, backendFile.FileRelLocation, backendFile.FileSize, backendFile.Status)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "阿里云分片上传文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: -1,
			Msg:  "阿里云分片上传文件成功！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}
}