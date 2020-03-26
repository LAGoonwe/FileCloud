package db

import (
	mydb "FileCloud/db/mysql"
	"fmt"
	"time"
)

//用户文件结构体
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
	SignedURL   string
}

//更新用户文件表
func OnUserFileUploadFinished(usename, filehash, filename string, filesize int64) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file (`user_name`,`file_sha1`,`file_name`,`file_size`,`upload_at`) values (?,?,?,?,?)")
	if err != nil {
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(usename, filehash, filename, filesize, time.Now())
	if err != nil {
		return false
	}
	return true
}

//批量获取用户文件信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare("select file_sha1,file_name,file_size,upload_at,last_update from tbl_user_file  where user_name = ? limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, ufile)
	}
	return userFiles, nil
}

//获取所有文件元数据信息
func GetAllFileMeta(limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare("select file_sha1,file_name,file_size,upload_at,last_update,user_name from tbl_user_file limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated, &ufile.UserName)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, ufile)
	}
	return userFiles, nil
}

//该db层方法提供给用户管理：删除系统用户的同时清空该用户的所有文件信息
func DeleteUserFileByUserAdmin(username string) {
	stmt, err := mydb.DBConn().Prepare(`DELETE FROM tbl_user_file WHERE user_name=?`)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec(username)
	if err != nil {
		fmt.Println(err.Error())
	}
}

//根据用户名查询用户文件表得到该用户的所有文件hash值
func GetAllFileHashByUsername(username string) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare("select file_sha1 from tbl_user_file where user_name = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, ufile)
	}
	return userFiles, nil
}
