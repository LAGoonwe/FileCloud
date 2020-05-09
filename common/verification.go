package common

const (
	// 校验用户名是否合法
	UserNameRegexp = "^[A-Za-z]+$"

	// 校验文件名是否合法
	FileNameRegexp = `^[^\/:*?"<>|]+$`

	// 校验文件hash是否合法
	FileHashRegexp = `^[A-Za-z0-9]+$`

	// 校验分页符是否合法
	LimitRegexp = "^[0-9]+$"
)
