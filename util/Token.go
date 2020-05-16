package util

/**
aud：标识token的接收者
exp：token过期时间，通常与Unix UTC时间做对比，过期后token无效
jti：自定义的id号
iat：签名发行时间
iss：签名发行者
nbf：token信息生效时间
sub：签名面向的用户
*/

import (
	"FileCloud/db"
	"errors"
	"fmt"
	"github.com/jwt-go"
	"time"
)

const SecretKey = "ZeusCloud"

/**
设置Payload结构体
*/
type jwtCustomClaims struct {
	jwt.StandardClaims

	//  追加信息（登录用户是否是管理员）
	Admin bool `json:"admin"`
}

/**
生成token
*/
func CreateToken(SecretKey []byte, username string, isAdmin bool) (tokenString string, err error) {
	claims := &jwtCustomClaims{
		StandardClaims: jwt.StandardClaims{
			// 设置token有效时间为24个小时
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Subject:   username,
		},
		Admin: false,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(SecretKey)
	return
}

/**
 * 解析 token,并且进行合法性验证
 */
func ParseToken(username, tokenSrt string, SecretKey []byte) (bool bool, err error) {
	var token *jwt.Token
	token, err = jwt.Parse(tokenSrt, func(*jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if token == nil {
		fmt.Println("Token对象为空")
		err = errors.New("Token对象为空")
		return false, err
	}
	// 验证token中的sub值是否与登录进的用户名一致
	claims := token.Claims

	if claims.(jwt.MapClaims)["sub"] != username {
		fmt.Println("Token对象错误")
		err = errors.New("Token对象错误")
		return false, err
	}

	// 中间件自我验证，包括是否过期等
	if !token.Valid {
		err = errors.New("Token失效")
		return false, err
	}

	// 与数据库存储的Token比较
	if tokenSrt != db.GetUserToken(username) {
		err = errors.New("Token与数据库不一致")
		return false, err
	}
	return true, nil
}

/**
测试
*/
//func main1() {
//	token, _ := CreateToken([]byte(SecretKey), "WinkiLee", true)
//	fmt.Println(token)
//
//	claims, err := ParseToken("WinkiLee",token, []byte(SecretKey))
//	if nil != err {
//		fmt.Println(" err :", err)
//	}
//	fmt.Println("claims:", claims)
//	//fmt.Println("claims uid:", claims.(jwt.MapClaims)["sub"])
//}
