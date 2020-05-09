package meta

import (
	dblayer "FileCloud/db"
)

// BackendFileMeta：文件信息结构
// 这个全局内存对象，如果删除了数据库文件需要重启才能清空
type BackendFile struct {
	UserName string
	FileSha1 string
	FileName string
	FileSize int64
	FileAbsLocation string
	FileRelLocation string
	Status int
}

var BackendFiles map[string]BackendFile

func init() {
	BackendFiles = make(map[string]BackendFile)
}

// 新增或更新文件信息
func UpdateBackendFileMeta(backendFileMeta BackendFile) {
	BackendFiles[backendFileMeta.FileSha1] = backendFileMeta
}

// 通过hash值获取文件信息对象
func GetBackendFileMeta(fileSha1 string) BackendFile {
	return BackendFiles[fileSha1]
}

// 通过hash值删除文件信息对象
func DeleteBackendFileMeta(fileSha1 string) {
	delete(BackendFiles, fileSha1)
}

// 通过sha1值从mysql中获取文件信息
func GetBackendFileFromDB(filesha1 string, username string) (BackendFile, error) {
	tBackendFile, err := dblayer.GetFileByFileSha1(filesha1, username)
	if err != nil {
		return BackendFile{}, err
	}
	backendFile := BackendFile{
		UserName:        tBackendFile.UserName,
		FileSha1:        tBackendFile.FileSha1,
		FileName:        tBackendFile.FileName,
		FileSize:        tBackendFile.FileSize,
		FileAbsLocation: tBackendFile.FileAbsLocation,
		FileRelLocation: tBackendFile.FileRelLocation,
		Status:          tBackendFile.Status,
	}
	return backendFile, nil
}