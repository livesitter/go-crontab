package master

import (
	"crontab/go-crontab/common"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
)

// 全局路由单例对象
var (
	G_apiServer *ApiServer
)

// 任务的http接口
type ApiServer struct {
	httpServer *http.Server
}

// 保存任务接口
// POST job={"name":"job1", "command":"echo hello", "cronExpr":"*****"}
func handleJobSave(w http.ResponseWriter, r *http.Request) {

	var (
		postedJob string
		err       error
		job       common.Job
		oldJob    *common.Job
		bytes     []byte
	)

	// 解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 取表单中的job字段
	postedJob = r.PostForm.Get("job")

	// 反序列化
	if err = json.Unmarshal([]byte(postedJob), &job); err != nil {
		goto ERR
	}

	// 保存任务到etcd
	if oldJob, err = JM.SaveJob(&job); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(bytes)
	}

	return

ERR:
	// 异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

// 删除任务接口
// POST /job/del  name=job1
func handleJobDel(w http.ResponseWriter, r *http.Request) {

	var (
		err    error
		name   string
		oldJob *common.Job
		bytes  []byte
	)

	// 解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 要删除的任务名
	name = r.PostForm.Get("name")

	// 删除任务
	if oldJob, err = JM.DeleteJob(name); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(bytes)
	}

	return
ERR:
	// 异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

// 任务列表接口
func handleJobList(w http.ResponseWriter, r *http.Request) {

	var (
		jobList []*common.Job
		err     error
		bytes   []byte
	)

	// 获取任务列表
	if jobList, err = JM.ListJobs(); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		w.Write(bytes)
	}

	return
ERR:
	// 异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

// 强杀任务接口
// POST /job/kill name=job1
func handleJobKill(w http.ResponseWriter, r *http.Request) {

	var (
		name  string
		err   error
		bytes []byte
	)

	// 解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 要杀掉的任务名
	name = r.PostForm.Get("name")

	// 杀掉任务
	if err = JM.KillJob(name); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		w.Write(bytes)
	}

	return
ERR:
	// 异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

// 初始化服务
func InitApiServer() (err error) {

	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
		// 静态文件根目录
		staticDir http.Dir
		// 静态文件回调
		staticHandler http.Handler
	)

	// 配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/del", handleJobDel)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)

	// 静态文件配置
	staticDir = http.Dir(Conf.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

	// 设置tcp监听端口
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(Conf.ApiPort)); err != nil {
		return
	}

	// 创建HTTP服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(Conf.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(Conf.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	// 路由对象赋值给全局单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	// 启动服务
	go httpServer.Serve(listener)

	return
}
