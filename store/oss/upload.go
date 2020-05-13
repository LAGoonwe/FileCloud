package oss

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aliyun-oss-go-sdk/oss"
	"log"
	"os"
	"strings"
)

// 上传字符串
func UploadString(objectName string, objectValue string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 指定存储类型为标准存储，缺省也为标准存储。
	storageType := oss.ObjectStorageClass(oss.StorageStandard)

	// 指定访问权限为公共读，缺省为继承bucket的权限
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	// 上传字符串
	err := bucket.PutObject(objectName, strings.NewReader(objectValue), storageType, objectAcl)
	if err != nil {
		log.Println("UploadString Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil

}


// 上传Byte数组
func UploadByte(objectName string, objectValue string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 指定存储类型为标准存储，缺省也为标准存储。
	storageType := oss.ObjectStorageClass(oss.StorageStandard)

	// 指定访问权限为公共读，缺省为继承bucket的权限
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	err := bucket.PutObject(objectName, bytes.NewReader([]byte("objectValue")), storageType, objectAcl)
	if err != nil {
		log.Println("DownLoadStream 获取文件流失败")
		log.Println(err.Error())
		return false, err
	}

	return true, nil
}



// 上传文件流
func UploadFileSteam(objectName string, fileName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 指定存储类型为标准存储，缺省也为标准存储。
	storageType := oss.ObjectStorageClass(oss.StorageStandard)

	// 指定访问权限为公共读，缺省为继承bucket的权限
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	// 读取本地文件
	file, err := os.Open(fileName)
	if err != nil {
		log.Println("UploadFileSteam Open File Failed")
		log.Println(err.Error())
		return false, err
	}
	defer file.Close()

	// 上传文件流
	err = bucket.PutObject(objectName, file, storageType, objectAcl)
	if err != nil {
		log.Println("UploadFileSteam Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 上传本地文件
func UploadLocalFile(objectName string, fileName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 指定存储类型为标准存储，缺省也为标准存储。
	storageType := oss.ObjectStorageClass(oss.StorageStandard)

	// 指定访问权限为公共读，缺省为继承bucket的权限
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	// 上传本地文件
	err := bucket.PutObjectFromFile(objectName, fileName, storageType, objectAcl)
	if err != nil {
		log.Println("UploadLocalFile Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 追加上传
// 文件不存在时，会创建一个可追加的文件；调用时，会向末尾追加内容
// 这里实现每次调用时追加一次内容
// 第一次追加可以指定文件元信息
func AppendFile(objectName string, appendValue string, startPos int64) (int64, bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return 0, false, errors.New("获取Bucket失败")
	}

	// nextPos的默认初始值为0
	// 返回值为下一次追加的位置
	nextPos, err := bucket.AppendObject(objectName, strings.NewReader(appendValue), startPos)
	if err != nil {
		log.Println("AppendFile Failed")
		log.Println(err.Error())
		return 0, false, err
	}
	return nextPos, true, nil
}



// 断点续传上传
func PartUpload(objectName string, filePath string, routineNum int) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	err := bucket.UploadFile(objectName, filePath, 100 * 1024, oss.Routines(routineNum), oss.Checkpoint(true, ""))
	if err != nil {
		log.Println("PartUpload Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 分片上传
// 这个可以上传较大的文件，因为价格的原因，还没做测试
// 分片上传和断点续传的区别：前者是一块块上传上去，后者是上传上去后再合并
func ComplexPartUpload(objectName string, filePath string, partNum int) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	chunks, err := oss.SplitFileByPartNum(filePath, partNum)
	file, err := os.Open(filePath)
	defer file.Close()

	// 指定存储类型为标准存储，缺省也为标准存储
	storageType := oss.ObjectStorageClass(oss.StorageStandard)

	// 1.初始化一个分片上传事件，指定存储类型为标准存储
	// 返回OSS创建的全局唯一的uploadId
	imur, err := bucket.InitiateMultipartUpload(objectName, storageType)
	uploadID := imur.UploadID
	fmt.Println(imur)
	fmt.Println(imur.UploadID)

	// 2.上传分片
	var parts []oss.UploadPart
	for _, chunk := range chunks {
		file.Seek(chunk.Offset, os.SEEK_SET)
		// 对每个分片调用UploadPart方法上传
		part, err := bucket.UploadPart(imur, file, chunk.Size, chunk.Number)
		if err != nil {
			log.Println("ComplexPartUpload UploadPart Failed")
			log.Println(err.Error())
		}
		parts = append(parts, part)
	}

	// 指定访问权限为公共读，缺省为继承bucket的权限
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	// 3.打印已上传的分片
	lsRes, err := bucket.ListUploadedParts(imur)
	if err != nil {
		log.Println("ComplexPartUpload bucket ListUploadedParts Failed")
		log.Println(err.Error())
		return false, err
	}
	// 打印已上传的分片
	fmt.Println(lsRes.UploadedParts)
	for _, upload := range lsRes.UploadedParts {
		fmt.Println("List PartNumber:  ", upload.PartNumber, ", ETag: " ,upload.ETag, ", LastModified: ", upload.LastModified)
	}
	// 根据objectName和UploadID生成InitiateMultipartUploadResult，然后列举所有已上传的分片，这种情况适用于已知objectName和UploadID的情况
	var imur_with_uploadid oss.InitiateMultipartUploadResult
	imur_with_uploadid.Key = objectName
	imur_with_uploadid.UploadID = uploadID
	// 列举已上传的分片
	lsRes, err = bucket.ListUploadedParts(imur_with_uploadid)
	if err != nil {
		log.Println("ComplexPartUpload ListUploadedParts Failed")
		log.Println(err.Error())
		return false, err
	}
	for _, upload := range lsRes.UploadedParts {
		fmt.Println("List PartNumber:  ", upload.PartNumber, ", ETag: ", upload.ETag, ", LastModified: ", upload.LastModified)
	}

	// 4.完成分片上传，指定访问权限为公共读
	cmur, err := bucket.CompleteMultipartUpload(imur, parts, objectAcl)
	if err != nil {
		log.Println("ComplexPartUpload CompleteMultipartUpload Failed")
		log.Println(err.Error())
		return false, err
	}
	fmt.Println(cmur)
	return true, nil
}



// 取消分片上传事件
func CancelPartUpload() {
}



// 列举分片上传事件
func ListUploadedChunks(objectName string, filePath string, uploadID int) {
}



// 进度条
// 实现后添加到上传方法中
// 定义进度条监听器。
type OssProgressListener struct {
}
// 定义进度变更事件处理函数
func (listener *OssProgressListener) ProgressChanged(event *oss.ProgressEvent) {
	switch event.EventType {
	case oss.TransferStartedEvent:
		fmt.Printf("Transfer Started, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferDataEvent:
		fmt.Printf("\rTransfer Data, ConsumedBytes: %d, TotalBytes %d, %d%%.",
			event.ConsumedBytes, event.TotalBytes, event.ConsumedBytes*100/event.TotalBytes)
	case oss.TransferCompletedEvent:
		fmt.Printf("\nTransfer Completed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferFailedEvent:
		fmt.Printf("\nTransfer Failed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	default:
	}
}


