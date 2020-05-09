package oss

import (
	"errors"
	"github.com/aliyun-oss-go-sdk/oss"
	"io"
	"log"
	"os"
)



// 下载文件到流
func DownLoadStream(objectName string) (io.Reader, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return nil, errors.New("获取Bucket失败")
	}

	// 下载文件到流
	body, err := bucket.GetObject(objectName)
	if err != nil {
		log.Println("DownLoadStream Failed")
		log.Println(err.Error())
		return nil, err
	}

	return body, nil
}



// 下载文件到流，再把流数据存入文件
func DownLoadLocalFileStream(objectName string, localFileName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 下载文件到本地文件流
	body, err := bucket.GetObject(objectName)
	if err != nil {
		log.Println("DownLoadLocalFile 下载文件到本地文件流失败")
		log.Println(err.Error())
		return false, err
	}
	defer body.Close()

	// os.O_WRONLY 只写模式打开文件，os.O_CREATE 如果文件不存在则创建文件
	file, err := os.OpenFile(localFileName, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Println("DownLoadLocalFile OpenFile Fail")
		log.Println(err.Error())
		return false, err
	}
	defer file.Close()

	io.Copy(file, body)
	return true, nil
}



// 范围下载：如果仅需要文件中的部分数据，您可以使用范围下载，下载指定范围内的数据
func DownLoadRangeFile(objectName string, start int64, end int64) (io.Reader, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return nil, errors.New("获取Bucket失败")
	}

	body, err := bucket.GetObject(objectName, oss.Range(start, end))
	if err != nil {
		log.Println("DownLoadRangeFile GetObject Fail")
		log.Println(err.Error())
		return nil, err
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄露，导致请求无连接可用
	defer body.Close()

	return body, nil
}



// 下载到本地文件
func DownLoadLocalFile(objectName string, localFileName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	err := bucket.GetObjectToFile(objectName, localFileName)
	if err != nil {
		log.Println("DownLoadLocalFile Fail")
		log.Println(err.Error())
		return false, err
	}

	return true, nil
}



// 断点续传下载
func DownLoadPartsFile(objectName string, localFileName string, routineNum int) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 分片下载
	// 设置三个协程并发下载分片，开启断点续传下载
	// Routines：指定分片下载的并发数。默认是1，即不使用并发下载
	// Checkpoint：指定是否开启断点续传下载功能以及设置Checkpoint文件。默认关闭断点续传下载功能
	// ss.Checkpoint(true, "")， 表示开启断点续传下载功能，并且Checkpoint文件为与本地文件同目录下的file.cp
	err := bucket.DownloadFile(objectName, localFileName, 10 * 1024, oss.Routines(routineNum), oss.Checkpoint(true, ""))
	if err != nil {
		log.Println("DownLoadPartsFile 断点续传下载文件失败")
		return false, err
	}
	return true, nil
}



//  文件压缩下载
// localFileName文件名后缀需要带上.gzip
func DownLoadGZIPFile(objectName string, localFileName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 文件压缩下载
	err := bucket.GetObjectToFile(objectName, localFileName, oss.AcceptEncoding("gzip"))
	if err != nil {
		log.Println("DownLoadPartsFile 断点续传下载文件失败")
		return false, err
	}
	return true, nil
}
