/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 16:04
 */

package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gowebsocket/helper"
	"gowebsocket/models"
	"net/http"
	"time"
)

var (
	clientManager = NewClientManager() // 管理者
	appIds        = []uint32{101, 102} // 全部的平台
	groupIds      = []uint32{101, 102} // 全部的机房


	serverIp   string
	serverPort string
)

func GetAppIds() []uint32 {

	return appIds
}

func GetServer() (server *models.Server) {
	server = models.NewServer(serverIp, serverPort)

	return
}

func IsLocal(server *models.Server) (isLocal bool) {
	if server.Ip == serverIp && server.Port == serverPort {
		isLocal = true
	}

	return
}

func InAppIds(appId uint32) (inAppId bool) {

	for _, value := range appIds {
		if value == appId {
			inAppId = true

			return
		}
	}

	return
}

func InGroupIds(groupId uint32) (inGroupIds bool) {
    inGroupIds = false
	for _, value := range groupIds {
		if value == groupId {
			inGroupIds = true

			return
		}
	}

	return
}

// 启动程序
func StartWebSocket() {

	if err := initSnowFlake(); err != nil {
		fmt.Println("StartWebSocket 启动雪花算法失败， err ", err)

	}


	serverIp = helper.GetServerIp()

	webSocketPort := viper.GetString("app.webSocketPort")
	//这里绑定的是错误的IP
	//rpcPort := viper.GetString("app.rpcPort")
	//serverPort = rpcPort

	http.HandleFunc("/acc", wsPage)

	// 添加处理程序
	go clientManager.start()
	fmt.Println("WebSocket 启动程序成功", serverIp, webSocketPort)

	//这里在启动了http server 服务，是client到websocket服务的请求端口，配置是8089
	http.ListenAndServe(":"+webSocketPort, nil)  
}

func wsPage(w http.ResponseWriter, req *http.Request) {

	// 升级协议
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		fmt.Println("升级协议", "ua:", r.Header["User-Agent"], "referer:", r.Header["Referer"])
		return true
	}}).Upgrade(w, req, nil)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	fmt.Println("webSocket 建立连接:", conn.RemoteAddr().String())

	currentTime := uint64(time.Now().Unix())
	client := NewClient(conn.RemoteAddr().String(), conn, currentTime)

	go client.read()
	go client.write()

	// 用户连接事件，只是连接上来，但是没有注册
	clientManager.Register <- client
}
