/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 16:24
 */

package websocket

import (
	"fmt"
	"runtime/debug"

	"github.com/gorilla/websocket"
)

const (
	// 用户连接超时时间
	heartbeatExpirationTime = 6 * 60
	allocateProtectDuration = 10 //单位为秒

)

// 用户登录
type login struct {
	AppId  uint32
	UserId string
	Client *Client
	//add for cloudmobile
	IsCloudmobile bool
	Group         uint32
	Uuid          string
}

// 读取客户端数据
func (l *login) GetKey() (key string) {
	if l.IsCloudmobile {
		key = GetCloudMobileKey(l.Group, l.Uuid)
	} else {
		key = GetUserKey(l.AppId, l.UserId)
	}

	return
}

const ( //云手机状态 state取值范围
	Good = 0
	Busy = 1
)

// 用户连接
type Client struct {
	Addr          string          // 客户端地址
	Socket        *websocket.Conn // 用户连接
	Send          chan []byte     // 待发送的数据
	AppId         uint32          // 登录的平台Id app/web/ios
	UserId        string          // 用户Id，用户登录以后才有
	FirstTime     uint64          // 首次连接事件
	HeartbeatTime uint64          // 用户上次心跳时间
	LoginTime     uint64          // 登录时间 登录以后才有
	//add for cloudmobile
	IsCloudmobile bool
	Group         uint32             //云手机机房id
	Name          string             //云手机名字
	Uuid          string             //云手机Uuid
	Allocated     bool               //是否被分配出去
	AllocateTime  uint64             //分配的时间
	AllocateUid   uint32             //分配给哪个uid
	Channel       uint64             //分配的RTC频道
	ch            chan *AllocateInfo //用来通知gin框架分配结果的channel
}

type AllocateInfo struct { //云手机的分配结果
	Code           uint32
	Client         *Client
	Uuid           string
	Group          uint32
	Rtc_channel    uint64
	Signal_channel uint64
}

// 初始化
func NewClient(addr string, socket *websocket.Conn, firstTime uint64) (client *Client) {
	client = &Client{
		Addr:          addr,
		Socket:        socket,
		Send:          make(chan []byte, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}

	return
}

// 读取客户端数据
func (c *Client) GetKey() (key string) {
	if c.IsCloudmobile {
		key = GetCloudMobileKey(c.Group, c.Uuid)
	} else {
		key = GetUserKey(c.AppId, c.UserId)
	}

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
		close(c.Send)
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			fmt.Println("读取客户端数据 错误", c.Addr, err)

			return
		}

		// 处理程序
		fmt.Println("读取客户端数据 处理:", string(message))
		ProcessData(c, message)
	}
}

// 向客户端写数据
func (c *Client) write() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("write stop", string(debug.Stack()), r)

		}
	}()

	defer func() {
		clientManager.Unregister <- c
		c.Socket.Close()
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

			c.Socket.WriteMessage(websocket.TextMessage, message)
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
func (c *Client) Login(appId uint32, userId string, loginTime uint64, isCloudmobile bool) {
	if isCloudmobile {
		c.Group = appId
		c.Uuid = userId
	} else {
		c.AppId = appId
		c.UserId = userId
	}
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
	if c.UserId != "" || c.Uuid != "" {
		isLogin = true

		return
	}

	return
}
