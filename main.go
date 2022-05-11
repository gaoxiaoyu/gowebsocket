/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-25
* Time: 09:59
 */

package main

import (
	"fmt"
	"gowebsocket/lib/redislib"
	"gowebsocket/routers"
	"gowebsocket/servers/grpcserver"
	"gowebsocket/servers/task"
	"gowebsocket/servers/websocket"
	"gowebsocket/lib/database"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	//defer undo()
	zap.L().Info("replaced zap's global loggers")

	sugar := logger.Sugar()
	sugar.Infof("name is %s", "x")                            // 格式化输出
	sugar.Infow("this is a test log", "name", "x", "age", 20) // 第二个开始每一对是一个键值
	// 使用全局的 SugaredLogger
	zap.S().Info("this is a test log: pid=", os.Getpid())

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
	redislib.ExampleNewClient()
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
