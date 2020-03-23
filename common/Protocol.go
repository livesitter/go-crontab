package common

import "encoding/json"

// 定时任务
type Job struct {
	// 任务名
	Name string `json:"name"`
	// shell命令
	Command string `json:"command"`
	// cron表达式
	CronExpr string `json:"cronExpr"`
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

// 返回应答函数
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {

	var (
		response Response
	)

	response.Errno = errno
	response.Msg = msg
	response.Data = data

	resp, err = json.Marshal(response)

	return
}
