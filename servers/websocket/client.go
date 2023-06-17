/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 16:24
 */

package websocket

import (
	"fmt"
	"gowebsocket/helper"
	"gowebsocket/models"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// 用户连接超时时间
	heartbeatExpirationTime = 6 * 60
	allocateProtectDuration = 10 //单位为秒

)

// 用户登录
type login struct {
	AppId  string
	UserId string
	Client *Client
}

// 读取客户端数据
func (l *login) GetKey() (key string) {

	key = GetUserKey(l.AppId, l.UserId)

	return
}

const ( //云手机状态 state取值范围
	Good = 0
	Busy = 1
)

// 用户连接
type Client struct {
	ConnId        uint64               //连接唯一id
	Addr          string               // 客户端地址
	Socket        *websocket.Conn      // 用户连接
	Send          chan []byte          // 待发送的数据
	Done          chan bool            //表示连接已经完成
	Received      chan bool            //表示收协程收到消息
	AppId         string               // 登录的平台Id app/web/ios
	UserId        string               // 用户Id，用户登录以后才有
	FirstTime     uint64               // 首次连接事件
	HeartbeatTime uint64               // 用户上次心跳时间
	LoginTime     uint64               // 登录时间 登录以后才有
	VerifyTime    uint64               // 连接验证时间
	UniClientInfo models.UniClientInfo //首次登录时保存的用户信息
}

// 初始化
func NewClient(addr string, socket *websocket.Conn, firstTime uint64) (client *Client) {
	client = &Client{
		ConnId:        helper.GenUint64Id(),
		Addr:          addr,
		Socket:        socket,
		Send:          make(chan []byte, 100),
		Done:          make(chan bool, 100),
		Received:      make(chan bool, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}

	return
}

// 读取客户端数据
func (c *Client) GetKey() (key string) {

	key = GetUserKey(c.AppId, c.UserId)

	return
}

// 读取客户端数据
func (c *Client) read() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("write stop", string(debug.Stack()), r)
		}
	}()

	defer func() {
		fmt.Println("读取客户端数据 关闭send", c)
		c.Socket.Close()
		c.close()
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			fmt.Println("读取客户端数据 错误", c.Addr, err)
			clientManager.Unregister <- c
			return
		}

		// 处理程序
		fmt.Println("读取客户端数据 处理:", string(message))
		zap.S().Debugw("client --> server", " connID:", c.ConnId, " addr:", c.Addr, "message:", string(message))
		c.Received <- true
		ProcessData(c, message)
	}
}

// 向客户端写数据
func (c *Client) write() {
	pingTicker := time.NewTicker(time.Second * 30)
	timeoutTimer := time.NewTimer(time.Second * heartbeatExpirationTime)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("write stop", string(debug.Stack()), r)
		}
	}()

	defer func() {
		c.Socket.Close()
		pingTicker.Stop()
		timeoutTimer.Stop()
		fmt.Println("Client发送数据 defer", c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// 发送数据错误 关闭连接
				fmt.Println("Client发送数据 关闭连接", c.Addr, "ok", ok)
				return
			}

			zap.S().Debugw("client <-- server", " connID:", c.ConnId, " addr:", c.Addr, "message:", string(message))
			c.Socket.WriteMessage(websocket.TextMessage, message)
		case <-pingTicker.C:
			ping := models.Ping{
				Ts: time.Now().Unix(),
			}
			msg := models.BuildUniMessage(helper.GetOrderIdTime(), "ping", models.UniMsgVersion1Define, ping)
			c.Send <- msg

		case <-timeoutTimer.C:
			zap.S().Debugw("client::write, timeout timer expired", "connID", c.ConnId, "addr", c.Addr)
			c.Done <- true

		case <-c.Received:
			timeoutTimer.Reset(time.Second * heartbeatExpirationTime)

		case <-c.Done:
			zap.S().Debugw("client::write, done", "connID", c.ConnId, "addr", c.Addr)
			return
		}
	}
}

// 读取客户端数据
func (c *Client) SendMsg(msg []byte) {

	if c == nil {

		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SendMsg stop:", r, string(debug.Stack()))
		}
	}()

	c.Send <- msg
}

// 读取客户端数据
func (c *Client) close() {
	close(c.Send)
}

// 用户登录
func (c *Client) Login(appId string, userId string, loginTime uint64) {
	c.AppId = appId
	c.UserId = userId
	c.LoginTime = loginTime
	// 登录成功=心跳一次
	c.Heartbeat(loginTime)
}

// 用户心跳
func (c *Client) Heartbeat(currentTime uint64) {
	c.HeartbeatTime = currentTime

	return
}

// 心跳超时
func (c *Client) IsHeartbeatTimeout(currentTime uint64) (timeout bool) {
	if c.HeartbeatTime+heartbeatExpirationTime <= currentTime {
		timeout = true
	}

	return
}

// 是否登录了
func (c *Client) IsLogin() (isLogin bool) {

	// 用户登录了
	if c.UserId != "" {
		isLogin = true

		return
	}

	return
}
