/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 17:36
 */

package models

import (
	"fmt"
	"time"
)

const (
	heartbeatTimeout = 3 * 60 // 用户心跳超时时间
)

// 用户在线状态
type UserOnline struct {
	AccIp         string `json:"accIp"`         // acc Ip
	AccPort       string `json:"accPort"`       // acc 端口
	AppId         uint32 `json:"appId"`         // appId
	UserId        string `json:"userId"`        // 用户Id
	ClientIp      string `json:"clientIp"`      // 客户端Ip
	ClientPort    string `json:"clientPort"`    // 客户端端口
	LoginTime     uint64 `json:"loginTime"`     // 用户上次登录时间
	HeartbeatTime uint64 `json:"heartbeatTime"` // 用户上次心跳时间
	LogOutTime    uint64 `json:"logOutTime"`    // 用户退出登录的时间
	Qua           string `json:"qua"`           // qua
	DeviceInfo    string `json:"deviceInfo"`    // 设备信息
	IsLogoff      bool   `json:"isLogoff"`      // 是否下线
	IsCloudMobile bool   `json:"isCloudMobile"` //是否云手机\
	Group         uint32 `json:"group"`         //云手机机房id
	Uuid          string `json:"uuid"`          //云手机uuid
	Name          string `json:"name"`          //云手机name
	State         uint32 `json:"state"`         //云手机可用状态
}

/**********************  数据处理  *********************************/

// 用户登录
func UserLogin(accIp, accPort string, appId uint32, userId string, addr string, loginTime uint64, isCloudMobile bool) (userOnline *UserOnline) {

	userOnline = &UserOnline{
		AccIp:         accIp,
		AccPort:       accPort,
		AppId:         appId,
		UserId:        userId,
		ClientIp:      addr,
		LoginTime:     loginTime,
		HeartbeatTime: loginTime,
		IsLogoff:      false,
		IsCloudMobile: isCloudMobile,
	}

	return
}

func CloudMobileLogin(accIp, accPort string, group uint32, uuid string, addr string, loginTime uint64, isCloudMobile bool, name string, state uint32) (userOnline *UserOnline) {

	userOnline = &UserOnline{
		AccIp:         accIp,
		AccPort:       accPort,
		Group:         group,
		Uuid:          uuid,
		ClientIp:      addr,
		LoginTime:     loginTime,
		HeartbeatTime: loginTime,
		IsLogoff:      false,
		IsCloudMobile: isCloudMobile,
		Name:          name,
		State:         state,
	}

	return
}

// 用户心跳
func (u *UserOnline) Heartbeat(currentTime uint64) {

	u.HeartbeatTime = currentTime
	u.IsLogoff = false

	return
}

// 用户退出登录
func (u *UserOnline) LogOut() {

	currentTime := uint64(time.Now().Unix())
	u.LogOutTime = currentTime
	u.IsLogoff = true

	return
}

/**********************  数据操作  *********************************/

// 用户是否在线
func (u *UserOnline) IsOnline() (online bool) {
	if u.IsLogoff {
		if u.IsCloudMobile {
			fmt.Println("用户是否在线 云手机已经下线", u.Group, u.Uuid)
		} else {
			fmt.Println("用户是否在线 用户已经下线", u.AppId, u.UserId)
		}
		return
	}

	currentTime := uint64(time.Now().Unix())

	if u.HeartbeatTime < (currentTime - heartbeatTimeout) {
		if u.IsCloudMobile {
		    fmt.Println("用户是否在线 云手机心跳超时", u.Group, u.Uuid, u.HeartbeatTime)
    	}else {
	 	    fmt.Println("用户是否在线 用户心跳超时", u.AppId, u.UserId, u.HeartbeatTime)
	    }

		return
	}

	/*if u.IsLogoff {
		fmt.Println("用户是否在线 用户已经下线", u.AppId, u.UserId)
		return
	}*/

	return true
}

// 用户是否在本台机器上
func (u *UserOnline) UserIsLocal(localIp, localPort string) (result bool) {

	if u.AccIp == localIp && u.AccPort == localPort {
		result = true

		return
	}

	return
}
