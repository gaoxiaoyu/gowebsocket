/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 16:24
 */

package websocket

import (
	"encoding/json"
	"fmt"
	"gowebsocket/helper"
	"gowebsocket/lib/database"
	"gowebsocket/lib/log"
	"gowebsocket/models"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 连接管理
type ClientManager struct {
	Clients     map[*Client]bool            // 全部的连接,还没有注册成功的时候
	ClientsLock sync.RWMutex                // 读写锁
	Users       map[string]*Client          // 登录的用户 // appId+uuid，注册成功才加入到这里
	UserLock    sync.RWMutex                // 读写锁
	Register    chan *Client                // 连接连接处理
	Login       chan *login                 // 用户登录处理
	Unregister  chan *Client                // 断开连接处理程序
	Broadcast   chan []byte                 // 广播 向全部成员发送数据
	Unicast     chan *models.UnicastMessage // 单播 向指定成员发送数据
	Allocated   chan uint32                 //删除分配记录
}

func NewClientManager() (clientManager *ClientManager) {
	clientManager = &ClientManager{
		Clients:    make(map[*Client]bool),
		Users:      make(map[string]*Client),
		Register:   make(chan *Client, 1000),
		Login:      make(chan *login, 1000),
		Unregister: make(chan *Client, 1000),
		Broadcast:  make(chan []byte, 1000),
		Unicast:    make(chan *models.UnicastMessage, 1000),
		Allocated:  make(chan uint32, 1000),
	}

	return
}

// 获取用户key
func GetUserKey(appId, userId string) (key string) {
	key = fmt.Sprintf("%s_%s", appId, userId)

	return
}

/**************************  manager  ***************************************/

// 添加客户端
func (manager *ClientManager) AddClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	manager.Clients[client] = true
}

// 删除客户端
func (manager *ClientManager) DelClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	delete(manager.Clients, client)
}

// 添加用户
func (manager *ClientManager) AddUsers(key string, client *Client) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	manager.Users[key] = client
}

// 删除用户
func (manager *ClientManager) DelUsers(key string) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	delete(manager.Users, key)
}

func (manager *ClientManager) GetUserClient(appId, userId string) (client *Client) {
	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()
	key := GetUserKey(appId, userId)
	client = manager.Users[key]
	return
}

// 向全部成员(除了自己)发送数据
func (manager *ClientManager) sendAll(message []byte, ignore *Client) {
	for conn := range manager.Clients {
		if conn != ignore {
			conn.SendMsg(message)
		}
	}
}

// 用户建立连接事件
func (manager *ClientManager) EventRegister(client *Client) {
	manager.AddClients(client)

	fmt.Println("EventRegister 用户建立连接", client.Addr)

	// client.Send <- []byte("连接成功")
}

// 用户登录
func (manager *ClientManager) EventLogin(login *login) {
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	client := login.Client
	// 连接存在，在添加
	if _, ok := manager.Clients[login.Client]; ok {
		userKey := login.GetKey()
		manager.AddUsers(userKey, login.Client)
	}

	fmt.Println("EventLogin 普通用户登录", client.Addr, login.AppId, login.UserId)
	orderId := helper.GetOrderIdTime()
	SendUserMessageAll(login.AppId, login.UserId, orderId, models.MessageCmdEnter, "哈喽~")

}

// 用户断开连接
func (manager *ClientManager) EventUnregister(client *Client) {
	manager.DelClients(client)

	// 删除用户连接
	var userKey string

	userKey = GetUserKey(client.AppId, client.UserId)

	manager.DelUsers(userKey)

	// 清除redis登录数据
	// userOnline, err := cache.GetUserOnlineInfo(client.GetKey())
	// if err == nil {
	// 	userOnline.LogOut()
	// 	cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	// }

	if err := database.DB().Debug().Where("app_id = ? anduser_id = ?", client.AppId, client.UserId).Delete(&models.UserOnlineInDb{}).Error; err != nil {
		zap.S().Errorw("EventUnregister, delte user in db err", "err", err, "appid", client.AppId, "userid", client.UserId)
	}

	// 关闭 chan
	// close(client.Send)
	{
		fmt.Println("EventUnregister 用户断开连接", client.Addr, client.AppId, client.UserId)
		if client.UserId != "" {
			orderId := helper.GetOrderIdTime()
			SendUserMessageAll(client.AppId, client.UserId, orderId, models.MessageCmdExit, "用户已经离开~")
		}
	}

}

// 管道处理程序
func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.Register:
			// 建立连接事件
			manager.EventRegister(conn)

		case login := <-manager.Login:
			// 用户登录
			manager.EventLogin(login)

		case conn := <-manager.Unregister:
			// 断开连接事件
			manager.EventUnregister(conn)

		case message := <-manager.Broadcast:
			// 广播事件
			for conn := range manager.Clients {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
				}
			}
		case unicastmsg := <-manager.Unicast:
			if client := manager.GetUserClient(unicastmsg.AppId, unicastmsg.UserId); client != nil {
				client.SendMsg(unicastmsg.Data)
			} else {
				log.Errorw("ClientManager::Unicast, Unicast to user failed", "appid", unicastmsg.AppId, "uid", unicastmsg.UserId, "user exist?", false)
			}
		}
	}
}

/**************************  manager info  ***************************************/
// 获取管理者信息
func GetManagerInfo(isDebug string) (managerInfo map[string]interface{}) {
	managerInfo = make(map[string]interface{})

	managerInfo["clientsLen"] = len(clientManager.Clients)
	managerInfo["usersLen"] = len(clientManager.Users)
	managerInfo["chanRegisterLen"] = len(clientManager.Register)
	managerInfo["chanLoginLen"] = len(clientManager.Login)
	managerInfo["chanUnregisterLen"] = len(clientManager.Unregister)
	managerInfo["chanBroadcastLen"] = len(clientManager.Broadcast)

	if isDebug == "true" {
		clients := make([]string, 0)
		for client := range clientManager.Clients {
			clients = append(clients, client.Addr)
		}

		users := make([]string, 0)
		for key := range clientManager.Users {
			users = append(users, key)
		}

		managerInfo["clients"] = clients
		managerInfo["users"] = users
	}

	return
}

// 获取用户所在的连接
func GetUserClient(appId, userId string) (client *Client) {
	client = clientManager.GetUserClient(appId, userId)

	return
}

// 定时清理超时连接
func ClearTimeoutConnections() {
	currentTime := uint64(time.Now().Unix())

	for client := range clientManager.Clients {
		if client.IsHeartbeatTimeout(currentTime) {

			fmt.Println("用户心跳时间超时 关闭连接", client.Addr, client.UserId, client.LoginTime, client.HeartbeatTime)

			client.Socket.Close()
		}
	}
}

// 获取全部用户
func GetUserList() (userList []string) {

	userList = make([]string, 0)
	fmt.Println("获取全部用户")

	for _, v := range clientManager.Users {
		userList = append(userList, v.UserId)
	}

	return
}

// 全员广播
func AllSendMessages(appId string, userId string, data string) {
	fmt.Println("全员广播", appId, userId, data)

	ignore := clientManager.GetUserClient(appId, userId)
	clientManager.sendAll([]byte(data), ignore)
}

func UnicastToClient(appid, clientid string, cmdDesc string, msg interface{}) {
	log.Debugw("UnicastToClient", "clientid", clientid, "appid", appid, "cmd", cmdDesc, "msg", msg)
	uniMsg := models.PrepareUniMessage(helper.GetOrderIdTime(), cmdDesc, models.UniMsgVersion1Define, msg)

	jsonmsg, err := json.Marshal(uniMsg)
	if err != nil {
		log.Debugw("UnicastToClient, unimsg json marshal error", "err", err, "clientid", clientid, "appid", appid, "cmd", cmdDesc, "msg", msg)
		return
	}

	unitcast := models.UnicastMessage{
		AppId:  appid,
		UserId: clientid,
		Data:   jsonmsg,
	}

	clientManager.Unicast <- &unitcast
}
