/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 17:28
 */

package cache

import (
	"encoding/json"
	"fmt"
	"gowebsocket/lib/redislib"
	"gowebsocket/models"

	"github.com/go-redis/redis"
)

const (
	userOnlinePrefix        = "acc:user:online:" // 用户在线状态
	cloudMobileOnlinePrefix = "uuid:"            // 用户在线状态
	userOnlineCacheTime     = 24 * 60 * 60
)

/*********************  查询用户是否在线  ************************/
func getUserOnlineKey(userKey string) (key string) {
	key = fmt.Sprintf("%s%s", userOnlinePrefix, userKey)

	return
}

func getCloudMobileOnlineKey(uuid string) (key string) {
	key = fmt.Sprintf("%s%s", cloudMobileOnlinePrefix, uuid)
	return
}

func GetUserOnlineInfo(userKey string) (userOnline *models.UserOnline, err error) {
	redisClient := redislib.GetClient()

	key := getUserOnlineKey(userKey)

	data, err := redisClient.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			fmt.Println("GetUserOnlineInfo", userKey, err)

			return
		}

		fmt.Println("GetUserOnlineInfo", userKey, err)

		return
	}

	userOnline = &models.UserOnline{}
	err = json.Unmarshal(data, userOnline)
	if err != nil {
		fmt.Println("获取用户在线数据 json Unmarshal", userKey, err)

		return
	}

	fmt.Println("获取用户在线数据", userKey, "time", userOnline.LoginTime, userOnline.HeartbeatTime, "AccIp", userOnline.AccIp)

	return
}

// 设置用户在线数据
func SetUserOnlineInfo(userKey string, userOnline *models.UserOnline) (err error) {

	redisClient := redislib.GetClient()
	key := getUserOnlineKey(userKey)

	valueByte, err := json.Marshal(userOnline)
	if err != nil {
		fmt.Println("设置用户在线数据 json Marshal", key, err)

		return
	}

	_, err = redisClient.Do("setEx", key, userOnlineCacheTime, string(valueByte)).Result()
	if err != nil {
		fmt.Println("设置用户在线数据 ", key, err)

		return
	}

	cloudmobileKey := getCloudMobileOnlineKey(userKey)
	data := make(map[string]interface{})
	data["accIp"] = userOnline.AccIp
	data["accPort"] = userOnline.AccPort
	data["appId"] = userOnline.AppId
	data["clientIp"] = userOnline.ClientIp
	data["clientPort"] = userOnline.ClientPort
	data["loginTime"] = userOnline.LoginTime
	data["heartbeatTime"] = userOnline.HeartbeatTime
	data["logOutTime"] = userOnline.LogOutTime
	data["name"] = userOnline.Name

	err = redisClient.HMSet(cloudmobileKey, data).Err()
	if err != nil {
		fmt.Println("设置云手机在线数据错误 ", cloudmobileKey, err)
		return
	}
	err = redisClient.Expire(cloudmobileKey, userOnlineCacheTime*1000*1000).Err()
	if err != nil {
		fmt.Println("设置云手机在线数据超时时间错误 ", cloudmobileKey, err)
		return
	}
	return

}

/*
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
	RtcChannel    uint64 `json:"rtcchannel"`     //rtc channel
	SignalChannel uint64 'json:"signalchannel"'  //signal channel

}*/
