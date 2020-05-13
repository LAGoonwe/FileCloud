package db

import (
	mydb "FileCloud/db/mysql"
	"errors"
	"log"
)

type FileMeta struct {
	UserFileId int
	Meta string
	UploadAt string
	LastUpdate string
}



// 获取所有的元信息
func GetAllObjectMeta(limit int) ([]FileMeta, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select user_file_id,meta,upload_at,last_update from tbl_user_file_meta limit ?")
	if err != nil {
		log.Println("GetAllObjectMeta DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		log.Println("GetAllObjectMeta Query Failed")
		log.Println(err.Error())
		return nil, err
	}

	var fileMetas []FileMeta
	for rows.Next() {
		fileMeta := FileMeta{}
		err := rows.Scan(
			&fileMeta.UserFileId,&fileMeta.Meta,&fileMeta.UploadAt,&fileMeta.LastUpdate)
		if err != nil {
			log.Println("GetAllObjectMeta Scan Failed")
			log.Println(err.Error())
			continue
		}
		fileMetas = append(fileMetas, fileMeta)
	}
	return fileMetas, nil
}



// 将获取的文件元信息插入到数据库中
func InsertFileMeta(userFileId int, meta string) (int, error) {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file_meta (`user_file_id`,`meta`) values (?,?)")
	if err != nil {
		log.Println("InsertFileMeta DB Failed")
		log.Println(err.Error())
		return 0, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userFileId, meta)

	if err != nil {
		log.Println("InsertFileMeta EXEC Failed")
		log.Println(err.Error())
		return 0, err
	}

	// 获取文件元信息主键
	var id int
	stmt2, err := mydb.DBConn().Prepare(
		"select id from tbl_user_file_meta where user_file_id = ?")
	if err != nil {
		log.Println("InsertFileMeta DB2 Failed")
		log.Println(err.Error())
		return 0, err
	}
	err = stmt2.QueryRow(userFileId).Scan(&id)
	if err != nil {
		log.Println("InsertFileMeta Scan Failed")
		log.Println(err.Error())
		return 0, err
	}
	return id, nil
}



// 更新文件元信息到数据库中
func UpdateFileMeta(userFileId int, meta string) (int, error) {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_user_file_meta set meta = ? where user_file_id = ?")
	if err != nil {
		log.Println("UpdateFileMeta DB Failed")
		log.Println(err.Error())
		return 0, err
	}
	row, err := stmt.Exec(meta, userFileId)
	if err != nil {
		log.Println("UpdateFileMeta EXEC Failed")
		log.Println(err.Error())
		return 0, err
	}
	if rf, err := row.RowsAffected(); err == nil {
		if rf < 0 {
			log.Println("UpdateFileMeta No Data Update")
			return 0, errors.New("数据更新异常")
		}
		var id int
		stmt2, err := mydb.DBConn().Prepare(
			"select id from tbl_user_file_meta where user_file_id = ?")
		if err != nil {
			log.Println("InsertFileMeta DB2 Failed")
			log.Println(err.Error())
			return 0, err
		}
		err = stmt2.QueryRow(userFileId).Scan(&id)
		if err != nil {
			log.Println("InsertFileMeta Scan Failed")
			log.Println(err.Error())
			return 0, err
		}
		return id, nil
	}
	return 0, errors.New("数据更新异常")
}




