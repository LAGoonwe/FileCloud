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

	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "更改文件状态失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	_, err = dblayer.UpdateFileStatus(filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "更改文件状态失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: 1,
		Msg:  "更改文件状态成功",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 移动所选文件
// 管理员可以把文件从一个用户目录移动到另外一个用户的目录
// 会把被移动的用户的文件给删除
func MoveFile(w http.ResponseWriter, r *http.Request) {
	// 移动所选文件的位置
	// OSS目前不支持直接移动文件，需要把文件Copy到目的地，再删除源目的地的文件
	// 链接：https://help.aliyun.com/knowledge_detail/39622.html?spm=5176.11065259.1996646101.searchclickresult.21387cc2wKX0vs&aly_as=GMs3aQodB
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	}

	r.ParseForm()
	username := r.Form.Get("username")
	checkusername := r.Form.Get("checkusername")
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

	// CheckGlobalFileMeta能够判断对应filehash的文件是否存在
	backendFile, err := CheckGlobalFileMeta(filehash)
	filename := backendFile.FileName
	if err != nil {
		log.Println(err.Error())
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 判断移动到的用户文件夹下是否有同名文件
	result, err := dblayer.IsExistSameNameFile(username, filename)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
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
			Msg:  "移动到的用户文件夹下存在相同名称的文件，移动失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 判断移动到的用户文件夹下是否有相同内容的文件
	result, err = dblayer.IsExistSameContentFile(username, filehash)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
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
			Msg:  "移动到的用户文件夹下存在相同内容的文件，移动失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 先进行OSS端的文件移动
	oldObjectName := backendFile.FileRelLocation
	destObjectName := checkusername + "/" + filename
	fmt.Println(destObjectName)

	// 本地服务器文件路径
	oldLocalFileName := backendFile.FileAbsLocation
	destLocalFileName := cfg.UploadPath + checkusername + "/" + filename
	fmt.Println(destLocalFileName)
	_, err = oss.CopyFiles(oldObjectName, destObjectName)
	if err != nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "移动所选文件失败！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
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
		return
	}

	// 移动本地服务器文件，还没测试是否会删除源目的地的文件
	os.Rename(oldLocalFileName, destLocalFileName)

	// 把更新的数据刷新到全局变量中
	backendFile.FileRelLocation = destObjectName
	backendFile.FileAbsLocation = destLocalFileName
	meta.BackendFiles[filehash] = *backendFile
	fmt.Println(meta.BackendFiles[filehash])
	// 把修改后的数据插入到数据库中
	// 上传完毕插入数据库，这里要确保不会发生异常
	_, err = dblayer.OnBackendUserFileUploadFinished(
		backendFile.UserName, backendFile.FileSha1, backendFile.FileName, backendFile.FileAbsLocation, backendFile.FileRelLocation, backendFile.FileSize, backendFile.Status)

	resp := util.RespMsg{
		Code: 1,
		Msg:  "移动文件成功",
		Data: "",
	}
	w.Write(resp.JSONBytes())
}



// 复制文件
// 不删除被复制的用户的文件
func CopyFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		username := r.Form.Get("username")
		checkusername := r.Form.Get("checkusername")
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

		// CheckGlobalFileMeta能够判断对应filehash的文件是否存在
		backendFile, err := CheckGlobalFileMeta(filehash)
		filename := backendFile.FileName
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "复制所选文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 判断复制到的用户文件夹下是否有同名文件
		result, err := dblayer.IsExistSameNameFile(username, filename)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "复制所选文件失败！",
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
				Msg:  "复制到的用户文件夹下存在相同名称的文件，复制失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 判断复制到的用户文件夹下是否有相同内容的文件
		result, err = dblayer.IsExistSameContentFile(username, filehash)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "复制所选文件失败！",
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
				Msg:  "复制到的用户文件夹下存在相同内容的文件，复制失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 先进行OSS端的文件移动
		oldObjectName := backendFile.FileRelLocation
		destObjectName := checkusername + "/" + filename
		fmt.Println(destObjectName)

		// 本地服务器文件路径
		destLocalFileName := cfg.UploadPath + checkusername + "/" + filename
		fmt.Println(destLocalFileName)
		_, err = oss.CopyFiles(oldObjectName, destObjectName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "复制所选文件失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 把更新的数据刷新到全局变量中
		backendFile.FileRelLocation = destObjectName
		backendFile.FileAbsLocation = destLocalFileName
		meta.BackendFiles[filehash] = *backendFile
		fmt.Println(meta.BackendFiles[filehash])
		// 把新的数据插入到数据库中
		// 上传完毕插入数据库，这里要确保不会发生异常
		_, err = dblayer.OnBackendUserFileUploadFinished(
			checkusername, backendFile.FileSha1, backendFile.FileName, destLocalFileName, destObjectName, backendFile.FileSize, 1)

		resp := util.RespMsg{
			Code: 1,
			Msg:  "复制文件成功",
			Data: "",
		}
		w.Write(resp.JSONBytes())
	}
}


// 修改文件内容（还未实现）
// 不同于重命名
func UpdateFileContent(w http.ResponseWriter, r *http.Request) {
	// 这里的修改文件是：用户重新上传文件，覆盖相同位置的原文件
	// 删除本地服务器文件和OSS文件，同时重新上传（相当于删除+上传结合）
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		// 上传的文件名要和原来的文件名一致
		file, head, err := r.FormFile("file")
		// 一定要传入filehash，确保用户修改的文件是存在的
		filehash := r.Form.Get("filehash")
		if err != nil {
			log.Println("UpdateFileContent Failed")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "修改文件内容失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		defer file.Close()

		fileMeta, err := CheckGlobalFileMeta(filehash)
		if err != nil {
			if err.Error() == "获取不到文件对象" {
				resp := util.RespMsg{
					Code: -1,
					Msg:  "修改的文件不存在！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			} else {
				resp := util.RespMsg{
					Code: -1,
					Msg:  "修改文件内容失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}
		}

		// 判断文件名是否一致
		filename := head.Filename
		if filename != fileMeta.FileName {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "用户修改后的文件名和原来的文件名不一致！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		localFileLocation := fileMeta.FileAbsLocation
		objectName := fileMeta.FileRelLocation
		fmt.Println(localFileLocation)
		fmt.Println(objectName)

		// 安全起见，创建临时存储文件
		preLocalFileLocation := cfg.PreUploadPath + username + "\\" + filename
		newPreFile, err := os.Create(preLocalFileLocation)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "修改文件内容失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		filesize, err := io.Copy(newPreFile, file)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "修改文件内容失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		newPreFile.Seek(0, 0)
		newFileSha1 := util.FileSha1(newPreFile)
		if newFileSha1 == fileMeta.FileSha1 {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "用户修改后的文件与原文内容一致！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		if !cfg.AsyncTransferEnable {
			// 先删除OSS上原来的文件
			_, err = oss.DeleteOneFile(objectName)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "修改文件内容失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}

			// 上传修改后的文件的临时文件
			err = oss.Bucket().PutObject(objectName, newPreFile)
			if err != nil {
				log.Println(objectName)
				log.Println(err.Error())
				resp := util.RespMsg{
					Code: -1,
					Msg:  "修改文件内容失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}
		} else {
			// 写入异步转移任务队列
		}

		// 删除本地服务器上的文件
		err = os.Remove(localFileLocation)
		if err != nil {
			log.Println("本地服务器删除文件失败")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "修改文件内容失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		newFile, err := os.Create(localFileLocation)
		if err != nil {
			log.Println("本地服务器创建文件失败")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "修改文件内容失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		_, err = io.Copy(newFile, newPreFile)
		if err != nil {
			log.Println("本地服务器拷贝文件失败")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "修改文件内容失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 删除临时文件
		//os.Remove(newPreFile)

		fileMeta.FileSha1 = newFileSha1
		fileMeta.FileSize = filesize
		meta.BackendFiles[filehash] = *fileMeta
		resp := util.RespMsg{
			Code: 1,
			Msg:  "修改文件内容成功！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
}



// 更改文件存储类型（待实现）
func ChangeFileStore(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {

	}
}