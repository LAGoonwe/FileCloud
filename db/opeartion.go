package db

import (
	mydb "FileCloud/db/mysql"
	"log"
	"strconv"
)

type Operation struct {
	OperationTypeId int
	MetaId          int
	UserName        string
	UserFileSha1    string
	Detail          string
	OperationTime   string
}

// 插入操作记录
func InsertOperation(typeOperationId int, metaId int, userFileSha1 string, userName, detail string) (int, error) {
	stmt, err := mydb.DBConn().Prepare(
		"insert into tbl_operation (`operation_type_id`,`user_name`,`meta_id`,`user_file_sha1`,`detail`) values (?,?,?,?,?)")
	if err != nil {
		log.Println("InsertOperation DB Failed")
		log.Println(err.Error())
		return 0, err
	}
	defer stmt.Close()

	row, err := stmt.Exec(typeOperationId, userName, metaId, userFileSha1, detail)
	resultId, _ := row.LastInsertId()
	resultIdInt, _ := strconv.Atoi(strconv.FormatInt(resultId, 10))
	if err != nil {
		log.Println("InsertOperation EXEC Failed")
		log.Println(err.Error())
		return 0, err
	}
	return resultIdInt, nil
}

// 根据文件元信息id获取操作记录
func GetOperationByMetaId(metaId int) ([]Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation where meta_id = ?")
	if err != nil {
		log.Println("GetOperationByMetaId DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	var operations []Operation
	rows, err := stmt.Query(metaId)
	for rows.Next() {
		operation := Operation{}
		err := rows.Scan(
			&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
		if err != nil {
			log.Println("GetAllOperation Scan Failed")
			log.Println(err.Error())
			continue
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

// 根据操作id查看操作记录
func GetOperationById(operationId int) (Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation where id = ?")
	if err != nil {
		log.Println("GetOperationById DB Failed")
		log.Println(err.Error())
		return Operation{}, err
	}
	defer stmt.Close()

	var operation Operation
	operation = Operation{}
	err = stmt.QueryRow(operationId).Scan(
		&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
	if err != nil {
		log.Println("GetOperationById Query Failed")
		log.Println(err.Error())
		return Operation{}, err
	}
	return operation, nil
}

// 获取所有的文件操作记录
func GetAllOperations(limit int) ([]Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation limit ?")
	if err != nil {
		log.Println("GetAllOperation DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	var operations []Operation
	rows, err := stmt.Query(limit)
	for rows.Next() {
		operation := Operation{}
		err := rows.Scan(
			&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
		if err != nil {
			log.Println("GetAllOperation Scan Failed")
			log.Println(err.Error())
			continue
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

// 获取指定文件的操作记录
func GetOperationsByUserFileId(userFileSha1 string) ([]Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation where user_file_sha1 = ?")
	if err != nil {
		log.Println("GetOperationByUserFileId DB Failed")
		log.Println(err.Error())
		return nil, err
	}

	var operations []Operation
	rows, err := stmt.Query(userFileSha1)
	for rows.Next() {
		operation := Operation{}
		err := rows.Scan(
			&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
		if err != nil {
			log.Println("GetAllOperation Scan Failed")
			log.Println(err.Error())
			continue
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

// 获取指定用户的操作记录
func GetOperationsByUserId(username string) ([]Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation where user_name = ?")
	if err != nil {
		log.Println("GetOperationByUserId DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	var operations []Operation
	rows, err := stmt.Query(username)
	for rows.Next() {
		operation := Operation{}
		err := rows.Scan(
			&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
		if err != nil {
			log.Println("GetOperationByUserId Scan Failed")
			log.Println(err.Error())
			continue
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

// 获取某个操作类型的操作记录
func GetOperationsByOperationType(operationTypeId int) ([]Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation where operation_type_id = ?")
	if err != nil {
		log.Println("GetOperationByOperationType DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	var operations []Operation
	rows, err := stmt.Query(operationTypeId)
	for rows.Next() {
		operation := Operation{}
		err := rows.Scan(
			&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
		if err != nil {
			log.Println("GetOperationByOperationType Scan Failed")
			log.Println(err.Error())
			continue
		}
		operations = append(operations, operation)
	}
	return operations, nil
}

// 获取某个时间段的操作记录
func GetOperationsByTime(startTime, endTime string) ([]Operation, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select operation_type_id,user_name,user_file_sha1,detail,operation_time from tbl_operation where operation_time >= ? and operation_time <= ?")
	if err != nil {
		log.Println("GetOperationByTime DB Failed")
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	var operations []Operation
	rows, err := stmt.Query(startTime, endTime)
	for rows.Next() {
		operation := Operation{}
		err := rows.Scan(
			&operation.OperationTypeId, &operation.UserName, &operation.UserFileSha1, &operation.Detail, &operation.OperationTime)
		if err != nil {
			log.Println("GetOperationByTime Scan Failed")
			log.Println(err.Error())
			continue
		}
		operations = append(operations, operation)
	}
	return operations, nil
}
