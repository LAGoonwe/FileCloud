package backendhandler

import (
	"FileCloud/common"
	dblayer "FileCloud/db"
	"FileCloud/meta"
	"errors"
	"log"
	"regexp"
)

// 判断请求接口的用户是否是系统管理员
func CheckUserStatus(username string) (bool, error) {
	user, err := dblayer.GetUserStatus(username)
	if err != nil {
		log.Println("CheckUserStatus Failed")
		log.Println(err.Error())
		return false, err
	}
	if user.Status != 7 {
		err := errors.New("Forbidden")
		return false, err
	}
	return true, nil
}



// 判断参数是否合法(通用)
func CheckParams(params map[string]string) (bool, error) {
	for k, v := range params {
		// 判断传入的用户名是否合法
		if k == "username" {
			length := len(v)
			result, _ := regexp.MatchString(common.UserNameRegexp, v)
			if length < 3 {
				log.Println("传入的用户名长度不够")
				err := errors.New("传入的用户名长度不够！")
				return false, err
				break
			}
			if !result {
				log.Println("传入的用户名非法")
				err := errors.New("传入的用户名非法！")
				return false, err
				break
			}
		}

		// 判断传入的文件名是否合法
		if k == "filename" {
			result, _ := regexp.MatchString(common.FileNameRegexp, v)
			if !result {
				log.Println("传入的文件名非法")
				err := errors.New("传入的文件名非法！")
				return false, err
				break
			}
		}

		// 判断传入的文件hash是否合法
		if k == "filehash" {
			length := len(v)
			result, _ := regexp.MatchString(common.FileHashRegexp, v)
			if length != 40 {
				log.Println("传入的文件hash长度不够")
				err := errors.New("传入的文件hash长度不够！")
				return false, err
				break
			}
			if !result {
				log.Println("传入的文件hash非法")
				err := errors.New("传入的文件hash非法！")
				return false, err
				break
			}
		}

		// 判断传入的分页符是否合法
		if k == "limit" {
			result, _ := regexp.MatchString(common.LimitRegexp, v)
			if !result {
				log.Println("传入的分页符非法")
				err := errors.New("传入的分页符非法！")
				return false, err
				break
			}
		}
	}
	return true, nil
}


// 操作时顺带将数据存入全局变量中，提升下一次的查找时间
func CheckGlobalFileMeta(filehash string) (*meta.BackendFile, error) {
	var backendfile meta.BackendFile
	if _, ok := meta.BackendFiles[filehash]; !ok {
		// 全局内存变量找不到内容就到数据库中查找
		backendUserFile, err := dblayer.GetLocalFile(filehash)

		// 找不到文件对象
		if backendUserFile.FileName == "" {
			log.Println("BackendUserFile Is Nil")
			// 随便返回一个空对象的地址
			return &meta.BackendFile{}, errors.New("获取不到文件对象")
		}

		backendfile = meta.BackendFile{}
		backendfile.UserName = backendUserFile.UserName
		backendfile.FileSha1 = backendUserFile.FileSha1
		backendfile.FileName = backendUserFile.FileName
		backendfile.FileSize = backendUserFile.FileSize
		backendfile.FileAbsLocation = backendUserFile.FileAbsLocation
		backendfile.FileRelLocation = backendUserFile.FileRelLocation
		meta.BackendFiles[backendfile.FileSha1] = backendfile

		if err != nil {
			log.Println("CheckGlobalFileMeta Failed")
			log.Println(err)
			return &meta.BackendFile{}, errors.New("获取文件信息时发生了错误")
		}
	} else {
		backendfile = meta.BackendFiles[filehash]
		log.Println(backendfile)
	}
	return &backendfile, nil
}
