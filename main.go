/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-25
* Time: 09:59
 */

package main

import (
	"fmt"
	"gowebsocket/helper"
	"gowebsocket/lib/database"
	"gowebsocket/lib/log"
	"gowebsocket/lib/redislib"
	"gowebsocket/routers"
	"gowebsocket/servers/grpcserver"
	"gowebsocket/servers/task"
	"gowebsocket/servers/websocket"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	//"gowebsocket/lib/log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	initConfig()

	initFile()
	initLogger()
	initZapLogerWithRotate()
	defer log.Sync()

	initRedis()
	database.InitDB()

	router := gin.Default()
	// 初始化路由
	routers.Init(router)
	routers.WebsocketInit()

	// 定时任务
	task.Init()

	// 服务注册
	task.ServerInit()

	go websocket.StartWebSocket()
	// grpc
	go grpcserver.Init()

	go open()

	httpPort := viper.GetString("app.httpPort")
	http.ListenAndServe(":"+httpPort, router)

}

// 初始化日志
func initFile() {
	// Disable Console Color, you don't need console color when writing the logs to file.
	gin.DisableConsoleColor()

	// Logging to a file.
	logFile := viper.GetString("app.logFile") //file path is current binary file path
	f, _ := os.Create(logFile)
	mydir, _ := os.Getwd()

	fmt.Println("log file path :", mydir, logFile)
	gin.DefaultWriter = io.MultiWriter(f)
}

// 初始化日志
func initLogger() {
	logFile := viper.GetString("app.processLogFile")
	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{logFile},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder, // INFO

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}.Build()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)
}

func initZapLogerWithRotate() {
	logPath := "./logs/"
	if err := helper.PathGuarantee(logPath); err != nil {
		panic(err)
	}

	_, processname := filepath.Split(os.Args[0])

	file, err := os.OpenFile(logPath+processname+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		panic(err)
	}

	logger := log.New(file, log.DebugLevel, zap.WithCaller(true), zap.Fields(zap.Any("pid", os.Getpid())), zap.AddCallerSkip(1))
	log.ResetDefault(logger)
}

func initConfig() {
	viper.SetConfigName("config/app")
	viper.AddConfigPath(".") // 添加搜索路径

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	fmt.Println("config app:", viper.Get("app"))
	fmt.Println("config redis:", viper.Get("redis"))

}

func initRedis() {
	redislib.NewClient()
}

/*
	type DBConfig struct {
		Name        string
		Dsn         string
		MaxIdleConn int
		MaxOpenConn int
	}

	[db.user]
dialect = mysql
dsn = root:123456@abcD@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local
max_idle_conn = 5
max_open_conn = 50
*/

func open() {

	time.Sleep(200 * time.Microsecond)

	httpUrl := viper.GetString("app.httpUrl")
	httpUrl = "http://" + httpUrl + "/home/index"

	fmt.Println("访问页面体验:", httpUrl)

	cmd := exec.Command("open", httpUrl)
	cmd.Output()
	//data, err := cmd.Output()
	//if err != nil {
	//	log.Fatalf("failed to call Output(): %v", err)
	//}
	//log.Printf("output: %s", data)
}
