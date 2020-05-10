package backendhandler

import "net/http"

// 原生的分块上传
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
	//if r.Method == http.MethodGet {
	//	w.Write([]byte("Forbidden"))
	//	return
	//} else if r.Method == http.MethodPost {
	//	r.ParseForm()
	//	username :=
	//}
}



// 上传文件分块
func UploadPart(w http.ResponseWriter, r *http.Request) {

}



// 通知上传并合并
func CompleteUpload(w http.ResponseWriter, r *http.Request) {

}



// 通知取消上传
func CancelUploadPart(w http.ResponseWriter, r *http.Request) {

}



// 查看分块上传状态
func MultipartUploadStatus(w http.ResponseWriter, r *http.Request) {

}