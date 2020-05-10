package util

import (
	cfg "FileCloud/config"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

func SendPhoneCode(code, phoneNumber string) error {
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", cfg.OSSAccesskeyID, cfg.OSSAccessKeySecret)

	request := dysmsapi.CreateSendSmsRequest()

	request.PhoneNumbers = phoneNumber
	request.SignName = "宙斯云盘"
	request.TemplateCode = "SMS_189762898"
	request.TemplateParam = "{\"code\":" + code + "}"

	response, err := client.SendSms(request)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	fmt.Printf("response is %#v\n", response)
	return nil
}
