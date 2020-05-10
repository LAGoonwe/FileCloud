package util

import (
	"fmt"
	"math/rand"
	"time"
)

/*
生成六位数注册验证码
*/
func CreateCaptcha() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}
