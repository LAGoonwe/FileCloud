package db

import (
	mydb "FileCloud/db/mysql"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// 用户文件结构体 tbl_user_file
type BackendUserFile struct {
	Id			    int
	UserName        string
	FileSha1        string
	FileName        string
	FileSize        int64
	FileAbsLocation string
	FileRelLocation string
	UploadAt        string
	LastUpdate      string
	Status          int
}



// 获取系统中所有用户的文件
func GetAllUserFiles(pageIndex, pageSize int) ([]BackendUserFile, error) {
	// 分页限制返回数量
	// 返回用户文件的所有信息：
	// 用户名，文件hash值，文件名，文件大小，文件绝对路径，文件相对路径（OSS），上传时间，最近更新时间，文件状态
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,upload_at,last_update,status from tbl_user_file limit ?,?")
	if err != nil {
		log.Println("GetAllUserFiles DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query((pageIndex-1)*pageSize, pageSize)
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


// 返回系统所有文件名--辅助搜索提示
func GetAllFilesExcPage() ([]BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_name from tbl_user_file")
	if err != nil {
		log.Println("GetAllFilesExcPage DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Println("GetAllFilesExcPage QUERY Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err := rows.Scan(&backendUserFile.FileName)
		if err != nil {
			log.Println("GetAllFilesExcPage Scan Failed")
			log.Println(err.Error())
			continue
		}
		backendUserFiles = append(backendUserFiles, backendUserFile)
	}
	return backendUserFiles, nil
}


// 根据名称检索文件
func GetFilesByName(name string) ([]BackendUserFile, error) {
	name = fmt.Sprintf("%s", "%"+name+"%")
	fmt.Println(name)
	// 分页限制返回数量
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,upload_at,last_update,status from tbl_user_file where CONCAT(`user_name`,`file_name`)  like ?")
	if err != nil {
		log.Println("GetFilesByName DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	fmt.Println(stmt)
	defer stmt.Close()

	rows, err := stmt.Query(name)
	fmt.Println(rows)
	if err != nil {
		log.Println("GetFilesByrName Query Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err := rows.Scan(&backendUserFile.UserName, &backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize, &backendUserFile.FileAbsLocation, &backendUserFile.FileRelLocation, &backendUserFile.UploadAt, &backendUserFile.LastUpdate, &backendUserFile.Status)
		if err != nil {
			log.Println("GetFilesByName Scan Failed")
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
	fmt.Println(backendUserFiles)
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

/**
冻结文件--更改文件status值
（暂时采取更改数据库信息的方法，后期考虑加入阿里云临时授权STS访问API以区别普通用户和管理员对文件的访问和操作）
1--文件可正常访问
0--文件被管理员冻结，用户不可访问
*/
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
		} else  {
			log.Println("GetLocalFileAllMeta QueryRow Failed")
			log.Println(err.Error())
			return BackendUserFile{}, err
		}
	}

	stmt2, err := mydb.DBConn().Prepare(
		"update tbl_user_file set status = ? where file_sha1 = ?")
	if err != nil {
		log.Println("GetLocalFileAllMeta Update1 Failed")
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
	result, err := stmt2.Exec(status, filehash)
	if err != nil {
		log.Println("UpdateFileStatus Update2 Failed")
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



// 判断文件是否已经被上传过
func IsFileHasUploaded(filesha1 string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare("" +
		"select id from tbl_user_file where file_sha1 = ?")
	if err != nil {
		log.Println("IsFileHasUploaded DB Failed")
		log.Println(err.Error())
		return false, err
	}
	var id int
	defer stmt.Close()

	err = stmt.QueryRow(filesha1).Scan(&id)
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
		"insert ignore into tbl_user_file (`user_name`,`file_sha1`,`file_name`,`file_size`,`file_abs_location`,`file_rel_location`,`status`) values (?,?,?,?,?,?,?)")
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
func QueryBackendUserFiles(username string, pageIndex, pageSize int) ([]BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,file_abs_location,file_rel_location,status,upload_at,last_update from tbl_user_file where user_name = ? limit ?,?")
	if err != nil {
		log.Println("QueryBackendUserFiles DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, (pageIndex-1)*pageSize, pageSize)
	if err != nil {
		log.Println("QueryBackendUserFiles Query Failed")
		log.Println(err.Error())
		return nil, err
	}

	var backendUserFiles []BackendUserFile
	for rows.Next() {
		backendUserFile := BackendUserFile{}
		err = rows.Scan(
			&backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize, &backendUserFile.FileAbsLocation,
			&backendUserFile.FileRelLocation, &backendUserFile.Status, &backendUserFile.UploadAt, &backendUserFile.LastUpdate)
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



// 根据用户名和文件名找到文件
func GetFileByUserNameAndFileName(username, filename string) (bool, BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,status from tbl_user_file where user_name = ? and file_name = ?")
	if err != nil {
		log.Println("GetFileByUserNameAndFileName DB Failed")
		log.Println(err.Error())
		return false, BackendUserFile{}, err
	}
	defer stmt.Close()

	backendFile := BackendUserFile{}
	err = stmt.QueryRow(username, filename).Scan(
		&backendFile.UserName, &backendFile.FileSha1, &backendFile.FileName, &backendFile.FileSize, &backendFile.FileAbsLocation, &backendFile.FileRelLocation, &backendFile.Status)
	if err == sql.ErrNoRows {
		return false, BackendUserFile{}, nil
	} else if err != nil {
		log.Println(err)
		return false, BackendUserFile{}, err
	}
	return true, backendFile, nil
}

// 更新追加上传后的文件大小和文件hash
func UpdateFileSizeAndFileHash(username, filename, filesha1 string, filesize int64) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_user_file set file_size = ? and file_sha1 = ? where user_name = ? and file_name = ?")
	if err != nil {
		log.Println("UpdateFileSizeAndFileHash DB Failed")
		log.Println(err.Error())
		return false, err
	}

	result, err := stmt.Exec(filesize, filesha1, username, filename)
	if err != nil {
		log.Println("UpdateFileSizeAndFileHash EXEC Failed")
		log.Println(err.Error())
		return false, err
	}
	if rf, err := result.RowsAffected(); err == nil {
		if rf < 0 {
			log.Println("DB UpdateFileStatus No Data Update")
			return false, errors.New("数据更新异常")
		}
	}
	return true, nil
}

// 查询当前用户下的Object是否存在
func GetFileByObjectName(objectName string) (BackendUserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select id,user_name,file_sha1,file_name,file_size,file_abs_location,file_rel_location,status from tbl_user_file where file_rel_location like ?")
	if err != nil {
		log.Println("IsExistObjectName DB Failed")
		log.Println(err.Error())
		return BackendUserFile{}, err
	}
	defer stmt.Close()

	var backendUserFile BackendUserFile
	backendUserFile = BackendUserFile{}
	err = stmt.QueryRow(objectName).Scan(&backendUserFile.Id, &backendUserFile.UserName, &backendUserFile.FileSha1, &backendUserFile.FileName, &backendUserFile.FileSize,
		&backendUserFile.FileAbsLocation, &backendUserFile.FileRelLocation, &backendUserFile.Status)
	if err == sql.ErrNoRows {
		log.Println("GetFileByObjectName Query Get No Data")
		return BackendUserFile{}, nil
	} else if err != nil {
		log.Println(err)
		return BackendUserFile{}, err
	}
	return backendUserFile, nil
}
