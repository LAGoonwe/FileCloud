package util

import (
	"encoding/json"
	"fmt"
	"log"
)

// RespMsg : http响应数据的通用结构
type RespMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// RespMsg2 : 分页逻辑专用
type RespMsg2 struct {
	Code             int         `json:"code"`
	Msg              string      `json:"msg"`
	Data             interface{} `json:"data"`
	FileTotal        int64       `json:"file_total"`
	UserTotal        int64       `json:"user_total"`
	AllFileMetaTotal int64       `json:"all_file_meta_total"`
}

func (resp *RespMsg2) JSONBytes2() []byte {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return r
}

// NewRespMsg : 生成response对象
func NewRespMsg(code int, msg string, data interface{}) *RespMsg {
	return &RespMsg{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

// JSONBytes : 对象转json格式的二进制数组
func (resp *RespMsg) JSONBytes() []byte {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return r
}

// JSONString : 对象转json格式的string
func (resp *RespMsg) JSONString() string {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return string(r)
}

// GenSimpleRespStream : 只包含code和message的响应体([]byte)
func GenSimpleRespStream(code int, msg string) []byte {
	return []byte(fmt.Sprintf(`{"code":%d,"msg":"%s"}`, code, msg))
}

// GenSimpleRespString : 只包含code和message的响应体(string)
func GenSimpleRespString(code int, msg string) string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, code, msg)
}
