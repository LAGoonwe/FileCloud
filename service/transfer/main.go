package main

import (
	"FileCloud/config"
	dblayer "FileCloud/db"
	"FileCloud/mq"
	"FileCloud/store/oss"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ProcessTransfer(msg []byte) bool {
	//解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	} else {
		rec_data, _ := json.Marshal(pubData)
		fmt.Println("成功接收消息：" + string(rec_data))
	}

	//根据临时存储文件路径，创建文件句柄
	filed, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//通过文件句柄将文件内容读出来并且上传到OSS
	err = oss.Bucket().PutObject(pubData.DestLocation, bufio.NewReader(filed))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//更新文件的存储路径到文件表
	suc := dblayer.UpdateFileLocation(pubData.FileHash, pubData.DestLocation)
	if !suc {
		return false
	}
	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中，开始监听转移任务队列...")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}
