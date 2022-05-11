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
	Good     = 0
	NotReady = 1
	Busy     = 2
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
	State         uint32             //云手机状态
	Allocated     bool               //是否被分配出去
	AllocateTime  uint64             //分配的时间
	AllocateUid   uint32             //分配给哪个uid
	RtcChannel    uint64             //分配的RTC频道
	SignalChannel uint64             //分配的信令频道
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

//设置云手机状态
func (c *Client) SetState(state uint32) (result bool, userId uint32) {
	result = true
    if c.State != state {
		switch c.State {
		case Good, NotReady:
			    //在空闲情况下，可以变为bad或者busy， 只是good-busy不应该发生 todo@：记录good -busy的发生次数
				//在bad情况下，可以变为good或者busy， 只是bad-busy 不应该发生，todo@：记录good -busy的发生次数
			    c.State = state
				zap.S().Info("SetState for: ", c.Addr, c.Uuid, "orig state: ", c.State, " to state:", state)

		case Busy:
				//在busy状态下，可以直接变为为NotReady，在保护时间之外，可以重新设置为idle，为了防止用户端手机掉新等故障导致云手机任务不再继续
				if state == Good {
					//检查分配时间
					currentTime := uint64(time.Now().Unix())
					if currentTime - c.AllocateTime >= allocateProtectDuration {
						c.State = state
						zap.S().Info("SetState for: ", c.Addr, c.Uuid, "from busy state to good state", "allocateTime:", c.AllocateTime, "allocateUid:", c.AllocateUid)
						userId = c.AllocateUid
					} else {
						zap.S().Info("SetState for: ", c.Addr, c.Uuid, " failed set from busy state to good state", "allocateTime:", c.AllocateTime, "now:", currentTime, "allocateUid:", c.AllocateUid)						
						result = false
					}
				} else {
				    //busy-bad的状态不需要保护时间，但是需要做错误处理，比如重新分配新的云手机 todo@: error recovery
					zap.S().Info("SetState for: ", c.Addr, c.Uuid, "from busy state to bad state, allocateUid:", c.AllocateUid)
					c.State = state
					userId = c.AllocateUid
				}
		}
	} else { //判断是否已经分配，此时state= good&idle，超过了分配保护时间，还是要充值分配状态
			if c.State == Good && c.Allocated {
				currentTime := uint64(time.Now().Unix())
				if currentTime - c.AllocateTime >= allocateProtectDuration {
					zap.S().Info("setstate for: ", c.Addr, c.Uuid, "from allocated state to idle state ", " allocatetime:", c.AllocateTime, " now:", currentTime, "allocateuid:", c.AllocateUid)
					userId = c.AllocateUid
				} else {
					result = false
					zap.S().Debug("setstate for: ", c.Addr, c.Uuid, "from allcated state to idle state failed for allocate protect,", "allocatetime:", c.AllocateTime, "now:", currentTime, "allocateuid:", c.AllocateUid)
				}
			}

	}


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
