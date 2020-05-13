package backendhandler

import (
	rPool "FileCloud/cache/redis"
	cfg "FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"FileCloud/store/oss"
	"FileCloud/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// 原生的分块上传
// TODO 分块上传后期扩展：可以新建文件来记录上传分块的情况
// 分块信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	// 适用于唯一的标志
	UploadID   string
	// 每个分块的大小
	ChunkSize  int
	// 分块的数量
	ChunkCount int
}



// 初始化分块上传
func InitialMultipartUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		filehash := r.Form.Get("filehash")
		filesize, err := strconv.Atoi(r.Form.Get("filesize"))
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		rConn := rPool.RedisPool().Get()
		defer rConn.Close()

		// 生成分块上传的初始化信息
		upInfo := MultipartUploadInfo{
			FileHash:   filehash,
			FileSize:   filesize,
			UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
			ChunkSize:  5 * 1024 * 1024,
			ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
		}

		// 初始化信息写入到Redis缓存
		// HSET 如果哈希表不存在，一个新的哈希表被创建并进行HSET操作，如果字段已经存在于哈希表中，旧值将被覆盖。
		// 新建字段返回1，旧值已存在被新值覆盖返回0
		_, err = rConn.Do("HSET", "MP_" + upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "初始化分块上传失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		_, err = rConn.Do("HSET", "MP_" + upInfo.UploadID, "filehash", upInfo.FileHash)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "初始化分块上传失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 向Redis缓存上传分块的信息
		_, err = rConn.Do("HSET", "MP_" + upInfo.UploadID, "filesize", upInfo.FileSize)
		if err != nil {
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "初始化分块上传失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 返回uploadID
		resp := util.RespMsg{
			Code: 1,
			Msg:  "初始化分块上传成功！",
			Data: upInfo.UploadID,
		}
		w.Write(resp.JSONBytes())
		return
	}
}



// 上传文件分块（上传单个分块）
// 分块序号的确定以及Request Body如何传输内容
func UploadPart(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		uploadID := r.Form.Get("uploadid")
		// 当前分块的序号（文件名）
		// 这里的问题是文件类型如何控制
		chunkIndex := r.Form.Get("index")

		rConn := rPool.RedisPool().Get()
		defer rConn.Close()

		// 还可以放入更多的关于分块的信息
		// 获得文件句柄，用于存储分块内容
		filepath := cfg.MultipartPath + uploadID + "/" + chunkIndex
		os.MkdirAll(path.Dir(filepath), 0744)
		file, err := os.Create(filepath)
		if err != nil {
			// 设置分块上传失败
			rConn.Do("HSET", "MP_" + uploadID, "chkidx_" + chunkIndex, -1)
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件分块失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		defer file.Close()

		buf := make([]byte, 1024 * 1024)
		for {
			// r.Body包含HTTP所要传输的内容，但并不是所有的报文都有主体
			n, err := r.Body.Read(buf)
			if err != nil {
				// 设置分块上传失败，直接结束
				rConn.Do("HSET", "MP_" + uploadID, "chkidx_" + chunkIndex, -1)
				resp := util.RespMsg{
					Code: -1,
					Msg:  "上传文件分块失败！",
					Data: "",
				}
				w.Write(resp.JSONBytes())
				return
			}
			file.Write(buf[:n])
		}

		// 更新Redis缓存状态，设置分块上传成功
		rConn.Do("HSET", "MP_" + uploadID, "chkidx_" + chunkIndex, 1)

		resp := util.RespMsg{
			Code: 1,
			Msg:  "上传文件分块成功！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
}



// 通知上传并合并
func CompleteUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		uploadID := r.Form.Get("uploadid")
		username := r.Form.Get("username")
		filehash := r.Form.Get("filehash")
		// filesize应该由前端来控制
		filesize := r.Form.Get("filesize")
		filename := r.Form.Get("filename")

		// 判断参数是否合法
		params := make(map[string]string)
		params["username"] = username
		params["filehash"] = filehash
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

		// 判断是否存在同名文件
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

		// 判断是否存在相同内容的文件
		result, err = dblayer.IsExistSameContentFile(username, filehash)
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

		rConn := rPool.RedisPool().Get()
		defer rConn.Close()

		// 通过uploadid查询Redis并判断是否所有分块上传完成
		// redis.Values返回一个[]interface{}，key & value 交替存入数组中
		data, err := redis.Values(rConn.Do("HGETALL", "MP_" + uploadID))
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "通知上传并合并失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		fmt.Println(data)

		totalCount := 0
		chunkCount := 0
		for i := 0; i < len(data); i += 2 {
			k := string(data[i].([]byte))
			v := string(data[i].([]byte))
			if k == "chunkcount" {
				totalCount, _ = strconv.Atoi(v)
			} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
				chunkCount++
			}
		}
		if totalCount != chunkCount {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "通知上传并合并失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 创建最终文件
		totalFileName := cfg.UploadPath + username + filename
		totalFile, err := os.Create(totalFileName)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "通知上传并合并失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 合并分块，读取分块文件内容，并合并到文件中
		for i := 0; i < len(data); i += 2 {
			k := string(data[i].([]byte))
			if strings.HasPrefix(k, "chkidx_") {
				// chunkIndex的默认开始位置为7
				chunkIndex := k[7:]
				fmt.Println(chunkIndex)
				// 获取分块文件的路径
				chunkFilePath := cfg.MultipartPath + uploadID + "/" + chunkIndex
				fmt.Println(chunkFilePath)
				file, err := os.Open(chunkFilePath)
				if err != nil {
					log.Println("CompleteUpload Failed1")
					log.Println(err.Error())
					resp := util.RespMsg{
						Code: -1,
						Msg:  "通知上传并合并失败！",
						Data: "",
					}
					w.Write(resp.JSONBytes())
					return
				}

				buf := make([]byte, 5 * 1024 * 1024)
				_, err = file.Read(buf)
				if err != nil {
					log.Println("CompleteUpload Failed2")
					log.Println(err.Error())
					resp := util.RespMsg{
						Code: -1,
						Msg:  "通知上传并合并失败！",
						Data: "",
					}
					w.Write(resp.JSONBytes())
					return
				}

				// 追加内容到最终文件
				end, err := totalFile.Seek(0, os.SEEK_END)
				if err != nil {
					log.Println("CompleteUpload Failed3")
					log.Println(err.Error())
					resp := util.RespMsg{
						Code: -1,
						Msg:  "通知上传并合并失败！",
						Data: "",
					}
					w.Write(resp.JSONBytes())
					return
				}
				// 追加内容到末尾
				_, err = totalFile.WriteAt(buf, end)
				if err != nil {
					log.Println("CompleteUpload Failed4")
					log.Println(err.Error())
					resp := util.RespMsg{
						Code: -1,
						Msg:  "通知上传并合并失败！",
						Data: "",
					}
					w.Write(resp.JSONBytes())
					return
				}
			}
		}

		// 创建临时文件
		totalPreFileName := cfg.PreUploadPath + username + filename
		totalPreFile, err := os.Create(totalPreFileName)
		if err != nil {
			log.Println("CompleteUpload Failed5")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "通知上传并合并失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		currentFileSize, err := io.Copy(totalPreFile, totalFile)
		currentFileSizeString := strconv.FormatInt(currentFileSize,10)
		// 文件大小不一致，证明文件分块过程有误
		if currentFileSizeString != filesize {
			fmt.Println(currentFileSize)
			fmt.Println(filesize)
			log.Println("CompleteUpload Failed6")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "通知上传并合并失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 计算文件hash值是否正确
		totalPreFile.Seek(0, 0)
		currentFileSha1 := util.FileSha1(totalPreFile)
		if currentFileSha1 != filehash {
			fmt.Println(currentFileSha1)
			fmt.Println(filehash)
			log.Println("CompleteUpload Failed6")
			log.Println(err.Error())
			resp := util.RespMsg{
				Code: -1,
				Msg:  "通知上传并合并失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 删除临时文件
		err = os.Remove(totalPreFileName)
		if err != nil {
			// 不影响主要业务
			log.Println("删除临时文件失败")
			log.Println(err.Error())
		}

		// 最后完成以后需要删除分块文件
		for i := 0; i < len(data); i += 2 {
			k := string(data[i].([]byte))
			if strings.HasPrefix(k, "chkidx_") {
				// chunkIndex的默认开始位置为7
				chunkIndex := k[7:]
				fmt.Println(chunkIndex)
				// 获取分块文件的路径
				chunkFilePath := cfg.MultipartPath + uploadID + "/" + chunkIndex
				err := os.Remove(chunkFilePath)
				if err != nil {
					log.Println("删除分块文件：", chunkIndex, "失败")
					log.Println(err.Error())
				}
			}
		}

		// 再走普通上传
		objectName := username + "/" + filename
		filesizeInt, _ := strconv.ParseInt(filesize, 10, 64)
		backendFile := meta.BackendFile{
			UserName:        username,
			FileSha1:        filehash,
			FileName:        filename,
			FileSize:        filesizeInt,
			FileAbsLocation: totalFileName,
			FileRelLocation: objectName,
			Status:          1,
		}

		// TODO 这块还需要完善 通过RabbitMQ实现异步上传
		if !cfg.AsyncTransferEnable {
			// OSS不需要判断上传的文件是否已存在OSS，因为数据库端和OSS保持一致，前面已做判断
			err = oss.Bucket().PutObject(objectName, totalFile)
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

		resp := util.RespMsg{
			Code: 1,
			Msg:  "通知上传并合并成功！",
			Data: "",
		}
		w.Write(resp.JSONBytes())
		return
	}
}



// 通知取消上传
// 流程：初始化分块上传信息->上传分块->通知分块合并
func CancelUploadPart(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		uploadID := r.Form.Get("uploadid")

		rConn := rPool.RedisPool().Get()
		defer rConn.Close()

		// 删除已存在的分块文件
		dirpath := cfg.MultipartPath + uploadID
		err := os.RemoveAll(path.Dir(dirpath))
		if err != nil {
			log.Println("删除本地分块文件失败")
			log.Println(err)
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件分块失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 删除Redis缓存情况
		_, err = rConn.Do("HDEL", "MP_" + uploadID)
		if err != nil {
			log.Println("删除Redis缓存信息失败")
			log.Println(err)
			resp := util.RespMsg{
				Code: -1,
				Msg:  "上传文件分块失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		// 待实现 更新mysql文件status（上传记录）
	}
}



// 查看分块上传状态
func MultipartUploadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
		return
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		uploadID := r.Form.Get("uploadid")

		rConn := rPool.RedisPool().Get()
		defer rConn.Close()

		// 获取分块初始化信息
		chunkCount, err := rConn.Do("HGET", "MP_" + uploadID, "chunkcount")
		if err != nil {
			log.Println(err)
			resp := util.RespMsg{
				Code: -1,
				Msg:  "查询分块上传状态失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		fmt.Println(chunkCount)

		filehash, err := rConn.Do("HGET", "MP_" + uploadID, "filehash")
		if err != nil {
			log.Println(err)
			resp := util.RespMsg{
				Code: -1,
				Msg:  "查询分块上传状态失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		fmt.Println(filehash)

		filesize, err := rConn.Do("HGET", "MP_" + uploadID, "filesize")
		if err != nil {
			log.Println(err)
			resp := util.RespMsg{
				Code: -1,
				Msg:  "查询分块上传状态失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		fmt.Println(filesize)

		// 检查分块上传状态是否有效
		data, err := redis.Values(rConn.Do("HGETALL", "MP_" + uploadID))
		if err != nil {
			log.Println(err)
			resp := util.RespMsg{
				Code: -1,
				Msg:  "查询分块上传状态失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}
		fmt.Println(data)

		// 分块那里还可以设置更多的信息
		var chunkInfos map[string]string
		chunkInfos = make(map[string]string)
		// 获取已上传的分块信息
		for i := 0; i < len(data); i += 2 {
			k := string(data[i].([]byte))
			v := string(data[i].([]byte))
			if strings.HasPrefix(k, "chkidx_") {
				chunkInfos[k] = v
			}
		}

		info := struct {
			ChunkCount interface{}
			Filehash interface{}
			Filesize interface{}
			ChunkInfos map[string]string
		} {
			ChunkCount: chunkCount,
			Filehash: filehash,
			Filesize: filesize,
			ChunkInfos: chunkInfos,
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "查询分块上传状态成功！",
			Data: info,
		}
		w.Write(resp.JSONBytes())
	}
}