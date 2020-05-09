package db

import (
	mydb "FileCloud/db/mysql"
	"database/sql"
	"errors"
	"log"
	"fmt"
)



// 用户文件结构体 tbl_user_file
type BackendUserFile struct {
	UserName    string
	FileSha1    string
	FileName    string
	FileSize    int64
	FileAbsLocation string
	FileRelLocation string
	UploadAt    string
	LastUpdate  string
	Status int
}



// 获取系统中所有用户的文件
func GetAllUserFiles(limit int) ([]BackendUserFile, error) {
	// 分页限制返回数量
	// 返回用户文件的所有信息：
	// 用户名，文件hash值，文件名，文件大小，文件绝对路径，文件相对路径（OSS），上传时间，最近更新时间，文件状态
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,upload_at,last_update,status from tbl_user_file limit ?")
	if err != nil {
		log.Println("GetAllUserFiles DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		log.Println("GetAllUserFiles QUERY Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err := rows.Scan(&backendUserFile.UserName, &backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize, &backendUserFile.FileAbsLocation, &backendUserFile.FileRelLocation, &backendUserFile.UploadAt, &backendUserFile.LastUpdate, &backendUserFile.Status)
		if err != nil {
			log.Println("GetAllUserFiles Scan Failed")
			log.Println(err.Error())
			continue
		}
		backendUserFiles = append(backendUserFiles, backendUserFile)
	}

	// 判断 backendUserFiles 是否为空
	// 当前策略是没有数据直接返回空即可
	//if (len(backendUserFiles) == 0) {
	//	err = errors.New("系统中没有用户文件或发生错误。")
	//	log.Println(err.Error())
	//	return nil, err
	//}
	return backendUserFiles, nil
}



// 根据用户名检索文件
func GetFilesByUserName(username string, limit int) ([]BackendUserFile, error) {
	username = fmt.Sprintf("%x", "%" + username + "%")
	// 分页限制返回数量
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,upload_at,last_update,status from tbl_user_file where user_name = ? limit ?")
	if err != nil {
		log.Println("GetFilesByUserName DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println("GetFilesByUserName Query Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err := rows.Scan(&backendUserFile.UserName, &backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize, &backendUserFile.FileAbsLocation, &backendUserFile.FileRelLocation, &backendUserFile.UploadAt, &backendUserFile.LastUpdate, &backendUserFile.Status)
		if err != nil {
			log.Println("GetFilesByUserName Scan Failed")
			log.Println(err.Error())
			continue
		}
		backendUserFiles = append(backendUserFiles, backendUserFile)
	}

	// 判断 backendUserFiles 是否为空
	// 当前策略是没有数据直接返回空即可
	//if (len(backendUserFiles) == 0) {
	//	err = errors.New("系统中没有用户文件或发生错误。")
	//	log.Println(err.Error())
	//	return nil, err
	//}
	return backendUserFiles, nil
}



// 根据文件名检索文件
func GetFilesByFileName(filename string, limit int) ([]BackendUserFile, error) {
	filename = fmt.Sprintf("%x", "%" + filename + "%")
	// 分页限制返回数量
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,upload_at,last_update,status from tbl_user_file where user_name = ? limit ?")
	if err != nil {
		log.Println("GetFilesByFileName DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(filename, limit)
	if err != nil {
		log.Println("GetFilesByFileName QUERY Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err := rows.Scan(&backendUserFile.UserName, &backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize, &backendUserFile.FileAbsLocation, &backendUserFile.FileRelLocation, &backendUserFile.UploadAt, &backendUserFile.LastUpdate, &backendUserFile.Status)
		if err != nil {
			log.Println("GetFilesByFileName Scan Failed")
			log.Println(err.Error())
			continue
		}
		backendUserFiles = append(backendUserFiles, backendUserFile)
	}

	// 判断 backendUserFiles 是否为空
	// 当前策略是没有数据直接返回空即可
	//if (len(backendUserFiles) == 0) {
	//	err = errors.New("系统中没有用户文件或发生错误。")
	//	log.Println(err.Error())
	//	return nil, err
	//}
	return backendUserFiles, nil
}



// 获取文件重要信息（通用）
func GetLocalFile(filesha1 string) (BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,status from tbl_user_file where file_sha1 = ?")
	if err != nil {
		log.Println("GetLocalFile DB Failed")
		log.Println(err.Error())
		return BackendUserFile{}, err
	}
	defer stmt.Close()

	var backendUserFile BackendUserFile
	backendUserFile = BackendUserFile{}
	row := stmt.QueryRow(filesha1)
	err = row.Scan(&backendUserFile.UserName, &backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize,
		&backendUserFile.FileAbsLocation, &backendUserFile.FileRelLocation, &backendUserFile.Status)
	if err == sql.ErrNoRows {
		log.Println("GetLocalFile Query Get No Data")
		return BackendUserFile{}, err
	} else if err != nil {
		log.Println("GetLocalFileMeta QueryRow Failed")
		log.Println(err.Error())
		return BackendUserFile{}, err
	}

	return backendUserFile, nil
}



// 修改文件状态
func UpdateFileStatus(filehash string) (BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select status from tbl_user_file where file_sha1 = ?")
	if err != nil {
		log.Println("UpdateFileStatus DB Failed")
		log.Println(err.Error())
		return BackendUserFile{}, err
	}
	defer stmt.Close()

	backendUserFile := BackendUserFile{}
	err = stmt.QueryRow(filehash).Scan(
		&backendUserFile.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			// 找不到数据
			log.Println("UpdateFileStatus No Data")
			return BackendUserFile{}, errors.New("数据不存在")
		} else {
			log.Println("GetLocalFileAllMeta QUERYROW Failed")
			log.Println(err.Error())
			return BackendUserFile{}, err
		}
	}

	stmt2, err := mydb.DBConn().Prepare(
		"update tbl_user_file set status = ? where file_sha1 = ?")
	if err != nil {
		log.Println("GetLocalFileAllMeta Update SQL1 Failed")
		log.Println(err.Error())
		return BackendUserFile{}, err
	}
	defer stmt2.Close()

	var status int
	if backendUserFile.Status == 1 {
		status = 0
		backendUserFile.Status = 1
	} else {
		status = 1
		backendUserFile.Status = 0
	}
	result, err := stmt.Exec(status, filehash)
	if err != nil {
		log.Println("UpdateFileStatus Update SQL2 Failed")
		log.Println(err.Error())
		return BackendUserFile{}, err
	}
	if rf, err := result.RowsAffected(); err == nil {
		if rf < 0 {
			log.Println("DB UpdateFileStatus No Data Update")
			return BackendUserFile{}, errors.New("数据更新异常")
		}
		return backendUserFile, nil
	}
	return BackendUserFile{}, errors.New("数据更新异常")
}



// 判断数据库中是否存在同名文件
func IsExistSameNameFile(username string, filename string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select id from tbl_user_file where user_name = ? and file_name = ? limit 1")
	if err != nil {
		log.Println("IsExistSameNameFile DB Failed")
		log.Println(err.Error())
		return false, err
	}
	var id int
	defer stmt.Close()

	//row, err := stmt.Query(username, filename)
	//if err != nil {
	//	log.Println("IsExistSameNameFile Query Failed")
	//	log.Println(err.Error())
	//	return false, err
	//} else if row == nil {
	//	return false, nil
	//}
	//return true, nil
	err = stmt.QueryRow(username, filename).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

// 判断数据库中是否存在内容相同的文件
func IsExistSameContentFile(username string, filesha1 string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select id from tbl_user_file where user_name = ? and file_sha1 = ? limit 1")
	if err != nil {
		log.Println("IsExistSameContentFile DB Failed")
		log.Println(err.Error())
		return false, err
	}
	var id int
	defer stmt.Close()

	//row, err := stmt.Query(username, filesha1)
	//fmt.Println(row)
	//if err != nil {
	//	log.Println("IsExistSameContentFile Query Failed")
	//	log.Println(err.Error())
	//	return false, err
	//} else if row == nil {
	//	return false, nil
	//}
	//return true, nil
	err = stmt.QueryRow(username, filesha1).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

// 上传完毕向用户文件表插入新的数据
func OnBackendUserFileUploadFinished(username, filesha1, filename, fileAbsLocation, fileRelLocation string, filesize int64, status int) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"insert into tbl_user_file (`user_name`,`file_sha1`,`file_name`,`file_size`,`file_abs_location`,`file_rel_location`,`status`) values (?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("OnUserFileUploadFinished DB Failed")
		log.Println(err.Error())
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filesha1, filename, filesize, fileAbsLocation, fileRelLocation, status)
	if err != nil {
		log.Println("OnUserFileUploadFinished EXEC Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}

// 通过filesha1获取文件对象
func GetFileByFileSha1(filesha1 string, username string) (*BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,status from tbl_user_file where file_sha1 = ? and user_name = ?")
	if err != nil {
		log.Println("GetFileByFileSha1 DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	backendFile := BackendUserFile{}
	err = stmt.QueryRow(filesha1, username).Scan(
		&backendFile.UserName, &backendFile.FileSha1, &backendFile.FileName, &backendFile.FileSize, &backendFile.FileAbsLocation, &backendFile.FileRelLocation, &backendFile.Status)
	if err != nil {
		log.Println("GetFileByFileSha1 QUERY Failed")
		log.Println(err.Error())
		return nil, err
	}
	return &backendFile, nil
}

// 更新文件绝路路径和文件相对路径
func UpdateFileName(filesha1, newFileName, fileAbsLocation, fileRelLocation string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_user_file set file_name=?, file_abs_location = ?, file_rel_location = ? where file_sha1 = ?")
	if err != nil {
		log.Println("UpdateFileName DB Failed")
		log.Println(err.Error())
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newFileName, fileAbsLocation, fileRelLocation, filesha1)
	if err != nil {
		log.Println("UpdateFileName EXEC Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}

// 批量获取用户文件信息（这里不展示filesha1，绝对路径）
// filesha1这种值不应该暴露给用户吗，本地服务器的绝对路径也是，存在安全风险
func QueryBackendUserFiles(username string, limit int) ([]BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_name,file_size,file_rel_location,status from tbl_user_file where user_name = ? limit ?")
	if err != nil {
		log.Println("QueryBackendUserFiles DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println("QueryBackendUserFiles Query Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err = rows.Scan(
			&backendUserFile.FileName, &backendUserFile.FileSize, &backendUserFile.FileRelLocation, &backendUserFile.Status)
		if err != nil {
			log.Println("QueryBackendUserFiles Scan Failed")
			log.Println(err.Error())
			return nil, err
		}
		backendUserFiles = append(backendUserFiles, backendUserFile)
	}
	return backendUserFiles, nil
}



// 删除数据库中的文件记录
func DeleteFile(filesha1 string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"delete from tbl_user_file where file_sha1 = ?")
	if err != nil {
		log.Println("DeleteFile DB Failed")
		log.Println(err.Error())
		return false, err
	}

	_, err = stmt.Exec(filesha1)
	if err != nil {
		log.Println("DeleteFile EXEC Failed")
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}