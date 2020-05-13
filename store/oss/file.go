package oss

import (
	"errors"
	"github.com/aliyun-oss-go-sdk/oss"
	"log"
	"net/http"
)

// 判断文件是否存在
func IsExistFile(objectName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 判断文件是否存在
	// 原则上是要先判断文件存不存在，再进行其他操作的
	isExist, err := bucket.IsObjectExist(objectName)
	if err != nil {
		log.Println("IsExistFile Failed")
		log.Println(err.Error())
		return false, err
	}
	return isExist, nil
}



// 设置文件访问权限
// 权限：
// 继承Bucket oss.ACLDefault：文件遵循存储空间的访问权限。
// 私有 oss.ACLPrivate：	文件的拥有者和授权用户有该文件的读写权限，其他用户没有权限操作该文件。
// 公共读 oss.ACLPublicRead：文件的拥有者和授权用户有该文件的读写权限，其他用户只有文件的读权限。
// 公共读写 oss.PublicReadWrite：所有用户都有该文件的读写权限。
func SetFileACL(objectName string, acl string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	var err error
	// 设置文件的访问权限
	if acl == "default" {
		err = bucket.SetObjectACL(objectName, oss.ACLDefault)
	} else if acl == "private" {
		err = bucket.SetObjectACL(objectName, oss.ACLPrivate)
	} else if acl == "publicread" {
		err = bucket.SetObjectACL(objectName, oss.ACLPublicRead)
	} else if acl == "publicreadwrite" {
		err = bucket.SetObjectACL(objectName, oss.ACLPublicReadWrite)
	} else {
		return false, errors.New("传入的权限码不正确！")
	}

	if err != nil {
		log.Println("SetFileACL Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 获取文件的访问权限
func GetFileACL(objectName string) (string, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return "", errors.New("获取Bucket失败")
	}

	aclRes, err := bucket.GetObjectACL(objectName)
	if err != nil {
		log.Println("GetFileACL Failed")
		log.Println(err.Error())
		return "", err
	}
	return aclRes.ACL, nil
}



// 获取文件元信息
// 文件元信息（Object Meta）是对上传到OSS的文件的属性描述，分为两种：HTTP标准属性（HTTP Headers）和 User Meta（用户自定义元信息）。
// 文件元信息可以在各种方式上传时或者拷贝文件时进行设置。
func GetFileMeta(objectName string) (http.Header, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return nil, errors.New("获取Bucket失败")
	}

	// 获取文件元信息
	props, err := bucket.GetObjectDetailedMeta(objectName)
	if err != nil {
		log.Println("GetFileMeta Failed")
		log.Println(err.Error())
		return nil, nil
	}
	return props, nil
}



// 设置文件元信息
// TODO 这里不实现，设置文件元信息可以在各种方式上传时或者拷贝文件时进行设置。
// 具体可设置的元信息看网页：https://help.aliyun.com/document_detail/88638.html?spm=a2c4g.11186623.6.960.3cbc5422SkmJs0
func SetFileMeta() {
}



// 修改文件元信息
func ModifyFileMeta(objectName string, props map[string]string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	var options []oss.Option
	for k, v := range props {
		options = append(options, oss.Meta(k, v))
	}
	err := bucket.SetObjectMeta(objectName, options...)
	if err != nil {
		log.Println("ModifyFileMeta Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 简单列举所有文件
// 还有更多的列举方式
func ListFiles() ([]string, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return nil, errors.New("获取Bucket失败")
	}

	// 列举所有文件
	marker := ""
	var allFileKeys []string
	for {
		lsRes, err := bucket.ListObjects(oss.Marker(marker))
		if err != nil {
			log.Println("ListObjects Failed")
			log.Println(err.Error())
			return nil, err
		}

		// 默认情况下一次返回100条记录
		for _, object := range lsRes.Objects {
			allFileKeys = append(allFileKeys, object.Key)
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}
	return allFileKeys, nil
}



// 删除单个文件
func DeleteOneFile(objectName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 删除单个文件
	err := bucket.DeleteObject(objectName)
	if err != nil {
		log.Println("DeleteOneFile Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, err
}



// 删除多个文件
func DeleteFiles(objectNames[] string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	// 不返回删除的结果
	_, err := bucket.DeleteObjects(objectNames)
	oss.DeleteObjectsQuiet(true)
	if err != nil {
		log.Println("DeleteFiles Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 同一存储空间复制文件
func CopyFiles(objectName string, destObjectName string) (bool, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return false, errors.New("获取Bucket失败")
	}

	_, err := bucket.CopyObject(objectName, destObjectName)
	if err != nil {
		log.Println("CopyFiles Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 获取文件下载外链
func GetDownLoadURL(objectName string) (string, error) {
	bucket := Bucket()
	if bucket == nil {
		log.Println("获取Bucket失败")
		return "", errors.New("获取Bucket失败")
	}

	downloadURL, err := Bucket().SignURL(objectName, oss.HTTPGet, 3600)
	if err != nil {
		log.Println("GetDownLoadURL")
		log.Println(err.Error())
		return "", err
	}
	return downloadURL, nil
}



// 更改文件存储类型
//func ChangeFileStore(objectName string, storeName string) (bool, error) {
//}

