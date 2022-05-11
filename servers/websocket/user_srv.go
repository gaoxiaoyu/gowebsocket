/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-30
* Time: 12:27
 */

package websocket

import (
	"errors"
	"fmt"
	"gowebsocket/lib/cache"
	"gowebsocket/models"
	"gowebsocket/common"
	"gowebsocket/helper"
	"gowebsocket/servers/grpcclient"
	"time"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
)

var node *snowflake.Node

// 查询所有用户
func UserList() (userList []string) {

	userList = make([]string, 0)
	currentTime := uint64(time.Now().Unix())
	servers, err := cache.GetServerAll(currentTime)
	if err != nil {
		fmt.Println("给全体用户发消息", err)

		return
	}

	for _, server := range servers {
		var (
			list []string
		)
		if IsLocal(server) {
			list = GetUserList()
		} else {
			list, _ = grpcclient.GetUserList(server)
		}
		userList = append(userList, list...)
	}

	return
}

// 查询用户是否在线
func CheckUserOnline(appId uint32, userId string) (online bool) {
	// 全平台查询
	if appId == 0 {
		for _, appId := range GetAppIds() {
			online, _ = checkUserOnline(appId, userId)
			if online == true {
				break
			}
		}
	} else {
		online, _ = checkUserOnline(appId, userId)
	}

	return
}

// 查询用户 是否在线
func checkUserOnline(appId uint32, userId string) (online bool, err error) {
	key := GetUserKey(appId, userId)
	userOnline, err := cache.GetUserOnlineInfo(key)
	if err != nil {
		if err == redis.Nil {
			fmt.Println("GetUserOnlineInfo", appId, userId, err)

			return false, nil
		}

		fmt.Println("GetUserOnlineInfo", appId, userId, err)

		return
	}

	online = userOnline.IsOnline()

	return
}

// 给用户发送消息
func SendUserMessage(appId uint32, userId string, msgId, message string) (sendResults bool, err error) {

	data := models.GetTextMsgData(userId, msgId, message)

	// TODO::需要判断不在本机的情况
	sendResults, err = SendUserMessageLocal(appId, userId, data)
	if err != nil {
		fmt.Println("给用户发送消息", appId, userId, err)
	}

	return
}

// 给本机用户发送消息
func SendUserMessageLocal(appId uint32, userId string, data string) (sendResults bool, err error) {

	client := GetUserClient(appId, userId)

	if client == nil {
		err = errors.New("用户不在线")

		return
	}

	// 发送消息
	client.SendMsg([]byte(data))
	sendResults = true

	return
}

// 给全体用户发消息
func SendUserMessageAll(appId uint32, userId string, msgId, cmd, message string) (sendResults bool, err error) {
	sendResults = true

	currentTime := uint64(time.Now().Unix())
	servers, err := cache.GetServerAll(currentTime)
	if err != nil {
		fmt.Println("给全体用户发消息", err)

		return
	}

	for _, server := range servers {
		if IsLocal(server) {
			data := models.GetMsgData(userId, msgId, cmd, message)
			AllSendMessages(appId, userId, data)
		} else {
			grpcclient.SendMsgAll(server, msgId, appId, userId, cmd, message)
		}
	}

	return
}

//func AllocateCloudMobile(userId uint32, ch chan *AllocateInfo) (sendResults bool, err error) {
//	//找到一个空闲的云手机
//	found, group, uuid := GetIdleCloudMobile()
//	if !found {
//		allocateInfo := &AllocateInfo{
//			Code: common.NoResource,
//		}
//		ch <- allocateInfo
//		return
//	}
//
//	//设为待分配状态
//	if !SetAllocateStatus(group, uuid) {
//		allocateInfo := &AllocateInfo{
//			Code: common.NoResource,
//		}
//		ch <- allocateInfo
//		return
//	}
//	//保存channle，等待分配结果的回复;
//	//分配rtc和signal channel给云手机
//	client := GetUserClient(group, uuid)
//	client.ch = ch
//	client.RtcChannel = getChannelId()
//	client.SignalChannel = getChannelId()
//
//	//发送AllocateReq给该手机
//	var (
//		code uint32
//		msg  string
//	)
//
//	code = common.OK
//	msg = common.GetErrorMessage(code, msg)
//	seq := helper.GetOrderIdTime()
//	cmd := "assign"
//
//	assignedReq := &models.AssignedReq{
//		Uid:            userId,
//		Rtc_channel:    client.RtcChannel,
//		Signal_channel: client.SignalChannel,
//	}
//
//	responseHead := models.NewResponseHead(seq, cmd, code, msg, assignedReq)
//
//	headByte, err := json.Marshal(responseHead)
//	if err != nil {
//		fmt.Println("处理数据 json Marshal", err)
//
//		return
//	}
//
//	client.SendMsg(headByte)
//	sendResults = true
//	fmt.Println("AllocateCloudMobile send", client.Addr, client.Group, client.Uuid, "cmd", cmd, "code", code, client.RtcChannel, client.SignalChannel)
//
//	return
//
//}

func AllocateCloudMobile(userId uint32) (result bool, rtcChannel uint64, signalChannel uint64) {

	//找到一个空闲的云手机
	found, group, uuid := GetIdleCloudMobile()
	if !found {
		fmt.Println("AllocateCloudMobile, 找不到可用空闲云手机 for uid ", userId)
		return
	}

	fmt.Println("AllocateCloudMobile, 找到空闲云手机 ", group, uuid)


	//设为待分配状态
	if !SetAllocateStatus(group, uuid, userId) {
		fmt.Println("AllocateCloudMobile, 设置云手机状态失败 ", group, uuid)
		return
	}

	//保存channle，等待分配结果的回复;
	//分配rtc和signal channel给云手机
	client := GetUserClient(group, uuid)
	rtcChannel = client.RtcChannel
	signalChannel = client.SignalChannel
	clientManager.AddAllocateRecord(userId, client)
	result = true

	fmt.Println("AllocateCloudMobile success ", client.Addr, client.Group, client.Uuid, client.RtcChannel, client.SignalChannel, userId)

	return

}


func RecyleCloudMobile(userId uint32) (result bool) {

	//找到一个空闲的云手机
	client := clientManager.GetAllocateRecord(userId)
	if client == nil {
		fmt.Println("RecyleCloudMobile, 找不到分配云手机记录 for uid ", userId)
		return
	}

	fmt.Println("RecyleCloudMobile, 找到空闲云手机 ", client.Group, client.Uuid)


	//设为待分配状态 
	if !ResetAllocateStatus(client.Group, client.Uuid) {
		fmt.Println("RecyleCloudMobile, 重置云手机分配状态失败 ", client.Group, client.Uuid)
		return
	}

	//查找rtc和signal channel，发送RecyleReq通知云手机回收
	var (
		code uint32
	    msg  string
	)

	code = common.OK
	msg = common.GetErrorMessage(code, msg)
	seq := helper.GetOrderIdTime()
	cmd := "recyle"

	recyleReq := &models.RecyleReq{
		Uid:            userId,
		Rtc_channel:    client.RtcChannel,
		Signal_channel: client.SignalChannel,
	}

	responseHead := models.NewResponseHead(seq, cmd, code, msg, recyleReq)

	headByte, err := json.Marshal(responseHead)
	if err != nil {
		fmt.Println("处理数据 json Marshal", err)

		return
	}

	client.SendMsg(headByte)
	
	clientManager.DelAllocateRecord(userId)
	
	result = true

	fmt.Println("AllocateCloudMobile success ", client.Addr, client.Group, client.Uuid, client.RtcChannel, client.SignalChannel, userId)

	return

}

func initSnowFlake() (err error) {
	var st time.Time
	var startTime string = "2021-04-28"
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	snowflake.Epoch = st.UnixNano() / 1000000
	node, err = snowflake.NewNode(1)
	return
}

func getChannelId() (channel uint64) {
	channel = uint64(node.Generate().Int64()) //利用雪花算法获取rtc和signal channel id
	fmt.Println("getChannelId  ", channel)
	return
}
