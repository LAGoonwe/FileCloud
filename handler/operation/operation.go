package operation

import (
	dblayer "FileCloud/db"
	filebackend "FileCloud/handler/filebackend"
	"FileCloud/util"
	"net/http"
	"strconv"
)



// 获取所有的操作记录
func GetAllOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		limitp := r.Form.Get("limit")
		limit, _ := strconv.Atoi(limitp)

		// 判断请求接口的用户是否是系统管理员
		_, err := filebackend.CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		allOperations, err := dblayer.GetAllOperations(limit)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取所有的操作记录失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取所有的操作记录成功！",
			Data: allOperations,
		}
		w.Write(resp.JSONBytes())
	}
}



// 获取指定文件的操作记录
func GetOperationsByUserFileId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		// 指定某个用户的某个文件
		checkusername := r.Form.Get("checkusername")
		filehash := r.Form.Get("filehash")

		// 判断请求接口的用户是否是系统管理员
		_, err := filebackend.CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		operations, err := dblayer.GetOperationsByUserFileId(checkusername, filehash)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取指定文件的操作记录失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取指定文件的操作记录成功！",
			Data: operations,
		}
		w.Write(resp.JSONBytes())
	}
}



// 获取某个用户的操作记录
func GetOperationsByUserId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		checkusername := r.Form.Get("checkusername")
		limitp := r.Form.Get("limit")
		limit, _ := strconv.Atoi(limitp)

		// 判断请求接口的用户是否是系统管理员
		_, err := filebackend.CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		operations, err := dblayer.GetOperationsByUserId(checkusername, limit)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取指定用户的操作记录失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取指定用户的操作记录成功！",
			Data: operations,
		}
		w.Write(resp.JSONBytes())
	}
}



// 获取某个时间段的操作记录
func GetOperationsByTime(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		startTime := r.Form.Get("start")
		endTime := r.Form.Get("end")
		limitp := r.Form.Get("limit")
		limit, _ := strconv.Atoi(limitp)

		// 判断请求接口的用户是否是系统管理员
		_, err := filebackend.CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		operations, err := dblayer.GetOperationsByTime(startTime, endTime, limit)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取指定时间段的操作记录失败！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取指定时间段的操作记录成功！",
			Data: operations,
		}
		w.Write(resp.JSONBytes())
	}
}


// 获取某个操作类型的操作记录
func GetOperationsByOperationType(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		limitp := r.Form.Get("limit")
		operationTypeId := r.Form.Get("operationid")
		operationTypeInt, _ := strconv.Atoi(operationTypeId)
		limit, _ := strconv.Atoi(limitp)

		// 判断请求接口的用户是否是系统管理员
		_, err := filebackend.CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		operations, err := dblayer.GetOperationsByOperationType(operationTypeInt, limit)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取某个操作类型的操作记录！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取某个操作类型的操作记录！",
			Data: operations,
		}
		w.Write(resp.JSONBytes())
	}
}



// 获取某个操作id对应的操作记录
func GetOperationsByOperationId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Write([]byte("Forbidden"))
	} else if r.Method == http.MethodPost {
		username := r.Form.Get("username")
		operationId := r.Form.Get("operationid")
		operationIdInt, _ := strconv.Atoi(operationId)

		// 判断请求接口的用户是否是系统管理员
		_, err := filebackend.CheckUserStatus(username)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "没有权限访问该接口！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		operation, err := dblayer.GetOperationById(operationIdInt)
		if err != nil {
			resp := util.RespMsg{
				Code: -1,
				Msg:  "获取某个操作类型的操作记录！",
				Data: "",
			}
			w.Write(resp.JSONBytes())
			return
		}

		resp := util.RespMsg{
			Code: 1,
			Msg:  "获取某个操作类型的操作记录！",
			Data: operation,
		}
		w.Write(resp.JSONBytes())
	}
}