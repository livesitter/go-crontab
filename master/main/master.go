package main

import (
	"crontab/go-crontab/master"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var (
	// 配置文件路径
	confFile string
)

// 解析命令行参数
func initArgs() {
	// 最终命令行执行命令：master -config ./master.json
	// 第一个参数绑定变量，这里指定为confFile
	// 第二个参数指定参数名，这里指定为config，也就是命令行-config后面紧跟着的
	// 第三个参数，默认值
	// 第四个参数，命令使用提示
	flag.StringVar(&confFile, "config", "./master.json", "指定master.json")

	// 执行解析
	flag.Parse()
}

// 初始化线程
func initEnv() {

	// 设置线程数量和cpu核心数量相等
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	var (
		err error
	)

	// 初始化命令行参数
	initArgs()

	// 初始化线程
	initEnv()

	// 加载配置
	if err = master.InitConfig(confFile); err != nil {
		fmt.Println(err)
		return
	}

	// 任务管理器
	if err = master.InitJobManager(); err != nil {
		fmt.Println(err)
		return
	}

	// 启动Api Http服务
	if err = master.InitApiServer(); err != nil {
		fmt.Println(err)
		return
	}

	// 维持主函数不退出
	for {
		time.Sleep(1 * time.Second)
	}

	// 正常退出
	return
}
