package oss

import (
	cfg "FileCloud/config"
	"fmt"
	"github.com/aliyun-oss-go-sdk/oss"
)

var ossCli *oss.Client

// Client : 创建oss client对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	ossCli, err := oss.New(cfg.OSSEndpoint,
		cfg.OSSAccesskeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	//fmt.Println("oss 对象生成成功")
	return ossCli
}

// Bucket : 获取bucket存储空间
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(cfg.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		//fmt.Println("oss 存储桶获取成功")
		return bucket
	}
	return nil
}

//DownloadUrl:临时授权下载
func DownloadURL(objName string) string {
	signedUrl, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedUrl
}

/**
其他功能
*/

/*
设定指定文件夹下文件生命周期：可用来做回收站功能
*/
func BuildLifecycleRule(bucketName string) {
	//表示前缀为test的对象（文件）距最后修改时间30天后过期
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}

	Client().SetBucketLifecycle(bucketName, rules)
}
