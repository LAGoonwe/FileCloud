package oss

import (
	"errors"
	"github.com/aliyun-oss-go-sdk/oss"
	"log"
)

// 创建存储空间
func CreateBucket(bucketName string) (bool, error) {
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return false, errors.New("与阿里云OSS连接失败")
	}

	// 创建存储空间
	// 命名规范：
	// 只能包括小写字母、数字和短横线（-）
	// 必须以小写字母或者数字开头和结尾
	// 长度必须在 3–63 字节之间\
	// 还可以继续指定存储类型
	err := client.CreateBucket(bucketName, oss.ACL(oss.ACLPublicReadWrite))
	if err != nil {
		log.Println("CreateBucket Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 列举所有存储空间
func ListAllBuckets([]string) ([]string, error) {
	// 存储空间按照字母顺序排列
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return nil, errors.New("与阿里云OSS连接失败")
	}

	// 列举存储空间
	// marker参数代表存储空间的名字，这里为空即为重头开始
	marker := ""
	var bucketNames []string
	for {
		lsRes, err := client.ListBuckets(oss.Marker(marker))
		if err != nil {
			log.Println("ListAllBuckets Failed")
			log.Println(err.Error())
			return nil, err
		}

		// 默认情况下返回100条记录
		for _, bucket := range lsRes.Buckets {
			bucketNames = append(bucketNames, bucket.Name)
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}
	return bucketNames, nil
}



// 判断存储空间是否存在
func IsExistBucket(bucketName string) (bool, error) {
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return false, errors.New("与阿里云OSS连接失败")
	}

	// 判断存储空间是否存在
	isExist, err := client.IsBucketExist(bucketName)
	if err != nil {
		log.Println("IsExistBucket Failed")
		log.Println(err.Error())
		return false, err
	}
	if isExist {
		return true, nil
	} else {
		return false, nil
	}
}



// 获取存储空间的信息
func GetBucketInfo(bucketName string) (map[string]interface{}, error)  {
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return nil, errors.New("与阿里云OSS连接失败")
	}

	// 获取存储空间的信息
	res, err := client.GetBucketInfo(bucketName)
	if err != nil {
		log.Println("GetBucketInfo Failed")
		log.Println(err.Error())
		return nil, err
	}

	var infoMaps map[string]interface{}
	infoMaps["Location"] = res.BucketInfo.Location
	infoMaps["CreationDate"] = res.BucketInfo.CreationDate
	infoMaps["ACL"] = res.BucketInfo.ACL
	infoMaps["Owner"] = res.BucketInfo.Owner
	infoMaps["StorageClass"] = res.BucketInfo.StorageClass
	infoMaps["RedundancyType"] = res.BucketInfo.RedundancyType
	infoMaps["ExtranetEndpoint"] = res.BucketInfo.ExtranetEndpoint
	infoMaps["IntranetEndpoint"] = res.BucketInfo.IntranetEndpoint
	return infoMaps, nil
}



// 设置存储空间的访问权限
func SetBucketACL(bucketNames string, acl string) (bool, error) {
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return false, errors.New("与阿里云OSS连接失败")
	}

	// 存储空间的访问权限（ACL）有以下三类：
	// 私有 oss.ACLPrivate：存储空间的拥有者和授权用户有文件的读写权限，其他用户没有权限操作文件。
	// 公共读 oss.ACLPublicRead：存储空间的拥有者和授权用户有文件的读写权限，其他用户只有读权限。
	// 公共读写 oss.ACLPublicReadWrite：所有用户都有文件的读写权限。请谨慎使用该权限
	var err error
	if acl == "private" {
		err = client.SetBucketACL(bucketNames, oss.ACLPrivate)
	} else if acl == "publicread" {
		err = client.SetBucketACL(bucketNames, oss.ACLPublicRead)
	} else if acl == "publicreadwrite" {
		err = client.SetBucketACL(bucketNames, oss.ACLPublicReadWrite)
	}
	if err != nil {
		log.Println("SetBucketACL Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



// 获取存储空间的访问权限
func GetBucketACL(bucketName string) (string, error) {
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return "", errors.New("与阿里云OSS连接失败")
	}

	// 获取存储空间的访问权限
	aclRes, err := client.GetBucketACL(bucketName)
	if err != nil {
		log.Println("GetBucketACL Failed")
		log.Println(err.Error())
		return "", err
	}
	return aclRes.ACL, nil
}



// 删除存储空间
func DeleteBucket(bucketName string) (bool, error) {
	client := Client()
	if client == nil {
		log.Println("与阿里云OSS连接失败")
		return false, errors.New("与阿里云OSS连接失败")
	}

	// 删除存储空间
	err := client.DeleteBucket(bucketName)
	if err != nil {
		log.Println("DeleteBucket Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}



/**
TODO
授权策略内容：
设置、获取、删除
 */

/**
TODO
生命周期
 */

/**
防盗链
 */
