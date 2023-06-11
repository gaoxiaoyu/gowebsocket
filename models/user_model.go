/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 17:36
 */

package models

import (
	"time"
)

const (
	heartbeatTimeout = 3 * 60 // 用户心跳超时时间
)

// 用户在线状态
type UserOnline struct {
	AppId         string `json:"appId"`         // appId
	UserId        string `json:"userId"`        // 用户Id
	ClientType    uint32 `json:"clientType"`    // clientType
	ClientId      string `json:"clientId"`      // 用户Id
	Name          string `json:"name"`          // name
	Platform      string `json:"platform"`      // 平台
	Ua            string `json:"ua"`            // ua
	ClientIp      string `json:"clientIp"`      // 客户端Ip
	ClientPort    string `json:"clientPort"`    // 客户端端口
	LoginTime     uint64 `json:"loginTime"`     // 用户上次登录时间
	HeartbeatTime uint64 `json:"heartbeatTime"` // 用户上次心跳时间
	LogOutTime    uint64 `json:"logOutTime"`    // 用户退出登录的时间
	AccIp         string `json:"accIp"`         // acc Ip
	AccPort       string `json:"accPort"`       // acc 端口
}

/**********************  数据处理  *********************************/

// 用户心跳
func (u *UserOnline) Heartbeat(currentTime uint64) {
	u.HeartbeatTime = currentTime
	return
}

// 用户退出登录
func (u *UserOnline) LogOut() {
	currentTime := uint64(time.Now().Unix())
	u.LogOutTime = currentTime
	return
}

/**********************  数据操作  *********************************/

// 用户是否在线
func (u *UserOnline) IsOnline() (online bool) {

	return u.LogOutTime != 0
}

// 用户是否在本台机器上
func (u *UserOnline) UserIsLocal(localIp, localPort string) (result bool) {

	if u.AccIp == localIp && u.AccPort == localPort {
		result = true

		return
	}

	return
}
