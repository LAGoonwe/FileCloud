package db

import (
	mydb "FileCloud/db/mysql"
	"database/sql"
	"fmt"
)

// OnFileUploadFinished : 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileabslocation, filerellocation string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_abs_location`,`file_rel_location`,`status`) values (?,?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement,err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileabslocation, filerellocation)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before----from CodeFile【db.file.OnFileUploadFinished】", filehash)
		}
		return true
	}

	return true
}

//用户文件信息重命名
func UpdateName(filename, filehash string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_user_file set`file_name`=? where  `file_sha1`=? ")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filename, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		//如果是新上传的文件影响数应该为0，如果是覆盖更新文件，影响数应该大于0
		if rf < 0 {
			fmt.Printf("更新文件名称失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}

//用户文件信息删除
func DeleteUserFile(filehash string) {
	stmt, err := mydb.DBConn().Prepare(`DELETE FROM tbl_user_file WHERE file_sha1=?`)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec(filehash)
	if err != nil {
		fmt.Println(err.Error())
	}
}

type TableFile struct {
	FileHash        string
	FileName        sql.NullString
	FileSize        sql.NullInt64
	FileAbsLocation sql.NullString
	FileRelLocation sql.NullString
	FileCreateAt    sql.NullTime
}

// 从mysql数据库查询获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_abs_location,file_rel_location,file_name,file_size,create_at from tbl_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileAbsLocation, &tfile.FileRelLocation, &tfile.FileName, &tfile.FileSize, &tfile.FileCreateAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return &tfile, nil
}

//UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set`file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		//如果是新上传的文件影响数应该为0，如果是覆盖更新文件，影响数应该大于0
		if rf < 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}

// 获取系统所有文件数
func GetFileNum() int64 {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT COUNT(*) FROM tbl_user_file;")
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	defer stmt.Close()
	var count int64
	err = stmt.QueryRow().Scan(&count)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return count
}
