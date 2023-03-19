/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-27
 * Time: 13:12
 */

package websocket

import (
	"encoding/json"
	"fmt"
	"gowebsocket/common"
	"gowebsocket/lib/cache"
	"gowebsocket/lib/database"
	"gowebsocket/models"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

// 用户登录
func LoginController(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{}) {

	autoRsp = true
	code = common.OK
	currentTime := uint64(time.Now().Unix())

	request := &models.Login{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ParameterIllegal
		fmt.Println("用户登录 解析数据失败", seq, err)

		return
	}

	fmt.Println("webSocket_request 用户登录", seq, "ServiceToken", request.ServiceToken)

	if request.UserId == "" || len(request.UserId) >= 20 {
		code = common.UnauthorizedUserId
		fmt.Println("用户登录 非法的用户", seq, request.UserId)

		return
	}

	if !InAppIds(request.AppId) {
		code = common.Unauthorized
		fmt.Println("用户登录 不支持的平台", seq, request.AppId)

		return
	}

	client.Login(request.AppId, request.UserId, currentTime, false)

	// 存储数据
	userOnline := models.UserLogin(serverIp, serverPort, request.AppId, request.UserId, client.Addr, currentTime, false)
	err := cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		fmt.Println("用户登录 SetUserOnlineInfo", seq, err)

		return
	}

	// 用户登录
	login := &login{
		AppId:  request.AppId,
		UserId: request.UserId,
		Client: client,
	}
	clientManager.Login <- login
	//todo@：这里准备写入mysql

	fmt.Println("用户登录 成功", seq, client.Addr, request.UserId)

	return
}

// 心跳接口
func HeartbeatController(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{}) {

	autoRsp = true
	code = common.OK
	currentTime := uint64(time.Now().Unix())

	request := &models.HeartBeat{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ParameterIllegal
		fmt.Println("心跳接口 解析数据失败", seq, err)
		zap.S().Info("Heartbeat decode error, from: ", client.Addr, "seq: ", seq)
		return
	}

	zap.S().Info("Heartbeat from: ", client.Addr, "seq: ", seq, "state:", request.State)
	if result, allocateUid := client.SetState(request.State); result && allocateUid != 0 {
		zap.S().Info("Heartbeat, setState success, restore allocate history for uid: ", allocateUid)
		if !ResetAllocateStatus(client.Group, client.Uuid) {
			fmt.Println("HeartbeatController, 重置云手机分配状态失败 ", client.Group, client.Uuid)
		}
		clientManager.Allocated <- allocateUid

	}

	data = &models.HeartBeatRsp{
		State: client.State,
	}

	fmt.Println("webSocket_request 心跳接口", client.AppId, client.UserId)

	if !client.IsLogin() {
		fmt.Println("心跳接口 用户未登录", client.AppId, client.UserId, seq)
		code = common.NotLoggedIn

		return
	}

	userOnline, err := cache.GetUserOnlineInfo(client.GetKey())
	if err != nil {
		if err == redis.Nil {
			code = common.NotLoggedIn
			fmt.Println("心跳接口 用户未登录", seq, client.AppId, client.UserId)

			return
		} else {
			code = common.ServerError
			fmt.Println("心跳接口 GetUserOnlineInfo", seq, client.AppId, client.UserId, err)

			return
		}
	}

	client.Heartbeat(currentTime)
	userOnline.Heartbeat(currentTime)
	err = cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		fmt.Println("心跳接口 SetUserOnlineInfo", seq, client.AppId, client.UserId, err)

		return
	}

	return
}

// register req
// {"seq":"2323","cmd":"register","data":{"uuid":"","state":0,"name":"xxx","group":0}}
func RegisterReqController(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{}) {
	zap.S().Info("RegisterReq from: ", client.Addr, "seq: ", seq)

	autoRsp = true
	code = common.OK
	currentTime := uint64(time.Now().Unix())

	request := &models.RegisterReq{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ParameterIllegal
		fmt.Println("云手机注册 解析数据失败", seq, err)
		return
	}

	fmt.Println("webSocket_request 云手机注册登录", seq, "uuid", request.Uuid, "group ", request.Group)

	if request.State > 3 { // name的规则也可以放这里匹配
		code = common.UnauthorizedUserId
		fmt.Println("云手机注册 错误的上报状态", seq, request.Uuid, request.State)
		return
	}

	if !InGroupIds(request.Group) {
		code = common.Unauthorized
		fmt.Println("云手机注册来自未配置的机房, seq:", seq, ",group:", request.Group)
		return
	}

	client.Login(request.Group, request.Uuid, currentTime, true)
	client.RtcChannel = getChannelId()
	client.SignalChannel = client.RtcChannel
	client.IsCloudmobile = true

	data = &models.RegisterRsp{
		Rtc_channel:    uint16(client.RtcChannel),
		Signal_channel: uint16(client.SignalChannel),
	}

	fmt.Println("云手机注册, 分配 rtc channel ", client.RtcChannel, " signal channel ", client.SignalChannel)

	// 存储数据
	fmt.Println("云手机注册 SetUserOnlineInfo start")

	userOnline := models.CloudMobileLogin(serverIp, serverPort, request.Group, request.Uuid, client.Addr, currentTime, true, request.Name, request.State)
	err := cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		fmt.Println("云手机注册 SetUserOnlineInfo", seq, err)
		return
	}

	if err := database.DB().Debug().Create(userOnline).Error; err != nil {
		fmt.Println("write cloud mobile to db failed ", err)
	}

	// 用户登录
	login := &login{
		IsCloudmobile: true,
		Group:         request.Group,
		Uuid:          request.Uuid,
		Client:        client,
	}

	clientManager.Login <- login

	fmt.Println("云手机注册 成功", seq, client.Addr, request.Uuid)

	return
}

// {"seq":"2323","cmd":"assign","data":{"code":200,"codeMsg":"Success", "uid":uid}}
func AssignedRspController(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{}) {

	autoRsp = false
	code = common.OK
	currentTime := uint64(time.Now().Unix())

	allocateInfo := &AllocateInfo{
		Code:           common.OK,
		Client:         client,
		Uuid:           client.Uuid,
		Group:          client.Group,
		Rtc_channel:    client.RtcChannel,
		Signal_channel: client.SignalChannel,
	}

	request := &models.AssignedRsp{}
	if err := json.Unmarshal(message, request); err != nil {
		allocateInfo.Code = common.ParameterIllegal
		fmt.Println("云手机分配 解析数据失败", seq, err)
		client.ch <- allocateInfo
		return
	}

	fmt.Println("webSocket_request 分配回复", seq, "uid", request.Userid, "code", request.Code)
	allocateInfo.Code = request.Code
	if request.Code == common.OK {
		client.State = Busy
	}

	client.AllocateTime = currentTime

	client.ch <- allocateInfo

	fmt.Println("云手机分配成功", seq, client.Addr, request.Userid, client.Uuid, client.RtcChannel, client.SignalChannel)

	return
}

// {"seq":"2323","cmd":"recyle","data":{"code":200,"codeMsg":"Success", "uid":uid}}
func RecyleRspController(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{}) {

	autoRsp = false

	request := &models.RecyleRsp{}
	if err := json.Unmarshal(message, request); err != nil {
		fmt.Println("云手机回收 解析数据失败", seq, err)
		zap.S().Info("RecyleRspController decode error, from: ", client, "msg: ", message, "seq: ", seq)
		return
	}

	fmt.Println("RecyleRspController 回收云手机回复", seq, "uid", request.Userid, "code", request.Code)
	zap.S().Info("RecyleRspController Recyle cloudmobile from: ", client.Addr, "uid", request.Userid, "code", request.Code)

	return
}
