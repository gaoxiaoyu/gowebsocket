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
	"gowebsocket/lib/database"
	"gowebsocket/models"
	"net"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

// 用户登录
func LoginController(client *Client, seq string, message []byte) (autoRsp bool, data interface{}) {

	autoRsp = true
	resp := &models.LoginRsp{}
	defer func() {
		if autoRsp {
			if resp.Code > 0 && resp.CodeMsg == "" {
				resp.CodeMsg = common.GetErrorMessage(resp.Code, resp.CodeMsg)
			}
			data = resp
		}
	}()

	currentTime := uint64(time.Now().Unix())

	request := &models.LoginReq{}
	if err := json.Unmarshal(message, request); err != nil {
		resp.Code = common.ParameterIllegal
		zap.S().Errorw("LoginController, decode LoginReq err", "err", err, "ClientId", client.UniClientInfo.ClientId)
		return
	}

	zap.S().Debugw("LoginController, LoginReq received", "userid", request.UserId, "appid", request.AppId, "from", client.Addr)

	if request.UserId == "" {
		resp.Code = common.UnauthorizedUserId
		zap.S().Errorw("LoginController, invailid UserId", "ClientId", client.UniClientInfo.ClientId)
		return
	}

	if !InAppIds(request.AppId) {
		resp.Code = common.Unauthorized
		fmt.Println("用户登录 不支持的平台", seq, request.AppId)
		return
	}

	client.Login(request.AppId, request.UserId, currentTime)
	clientIp, clientPort, err := net.SplitHostPort(client.Addr)
	if err != nil {
		zap.S().Errorw("LoginController, wrong client addr ", "client addr", client.Addr)
		resp.Code = common.ParameterIllegal
		return
	}

	userOnline := models.UserOnlineInDb{
		AppId:         request.AppId,
		UserId:        request.UserId,
		ClientType:    client.UniClientInfo.ClientType,
		ClientId:      client.UniClientInfo.ClientId,
		Name:          request.Name,
		Platform:      client.UniClientInfo.Platform,
		Ua:            client.UniClientInfo.Ua,
		ClientIp:      clientIp,
		ClientPort:    clientPort,
		LoginTime:     currentTime,
		HeartbeatTime: currentTime,
		LogOutTime:    0,
		AccIp:         serverIp,
		AccPort:       serverPort,
	}
	if err := database.DB().Debug().Clauses(clause.OnConflict{UpdateAll: true}).Create(&userOnline).Error; err != nil {
		resp.Code = common.ModelAddError
		zap.S().Errorw("LoginController, add user in db err", "err", err, "ClientId", client.UniClientInfo.ClientId)
		return
	}

	// 存储数据
	// userOnline := models.UserLogin(serverIp, serverPort, request.AppId, request.UserId, client.Addr, currentTime, false)
	// err := cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	// if err != nil {
	// 	resp.Code = common.ServerError
	// 	fmt.Println("用户登录 SetUserOnlineInfo", seq, err)
	// 	return
	// }

	// 用户登录
	login := &login{
		AppId:  request.AppId,
		UserId: request.UserId,
		Client: client,
	}
	clientManager.Login <- login
	zap.S().Infow("LoginController, user login sucess", "userid", request.UserId, "appid", request.AppId, "addr", client.Addr)
	resp.Code = common.OK

	return
}

// 心跳接口
func HeartbeatController(client *Client, seq string, message []byte) (autoRsp bool, data interface{}) {
	autoRsp = true
	resp := &models.HeartBeatRsp{
		RspCodeInfo: models.RspCodeInfo{
			Code: common.OK,
		},
	}
	defer func() {
		if autoRsp {
			if resp.Code > 0 && resp.CodeMsg == "" {
				resp.CodeMsg = common.GetErrorMessage(resp.Code, resp.CodeMsg)
			}
			data = resp
		}
	}()

	currentTime := uint64(time.Now().Unix())

	request := &models.HeartBeatReq{}
	if err := json.Unmarshal(message, request); err != nil {
		resp.Code = common.ParameterIllegal
		fmt.Println("心跳接口 解析数据失败", seq, err)
		zap.S().Info("HeartbeatController,  decode HeartBeatReq error, from: ", client.Addr, "seq: ", seq)
		return
	}

	zap.S().Info("HeartbeatController, receive heartbeat", "addr", client.Addr, "ClientId", client.UniClientInfo.ClientId, "state:", request.State)

	resp.State = request.State

	fmt.Println("webSocket_request 心跳接口", client.AppId, client.UserId)

	if !client.IsLogin() {
		fmt.Println("心跳接口 用户未登录", client.AppId, client.UserId, seq)
		resp.Code = common.NotLoggedIn
		return
	}

	// userOnline, err := cache.GetUserOnlineInfo(client.GetKey())
	// if err != nil {
	// 	if err == redis.Nil {
	// 		//code = common.NotLoggedIn
	// 		fmt.Println("心跳接口 用户未登录", seq, client.AppId, client.UserId)

	// 		return
	// 	} else {
	// 		//code = common.ServerError
	// 		fmt.Println("心跳接口 GetUserOnlineInfo", seq, client.AppId, client.UserId, err)

	// 		return
	// 	}
	// }

	client.Heartbeat(currentTime)
	// userOnline.Heartbeat(currentTime)
	// err = cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	// if err != nil {
	// 	//code = common.ServerError
	// 	fmt.Println("心跳接口 SetUserOnlineInfo", seq, client.AppId, client.UserId, err)

	// 	return
	// }

	if err := database.DB().Debug().Model(&models.UserOnlineInDb{}).Where("user_id = ?", client.UserId).Updates(models.UserOnlineInDb{HeartbeatTime: currentTime}).Error; err != nil {
		resp.Code = common.ModelAddError
		zap.S().Errorw("LoginController, update user heartbeat time in db err", "err", err, "userId", client.UserId)
		return
	}

	return
}

func SendToUserMsgReqController(client *Client, seq string, message []byte) (autoRsp bool, data interface{}) {
	autoRsp = true
	resp := &models.SendToUserMsgRsp{
		RspCodeInfo: models.RspCodeInfo{
			Code: common.OK,
		},
	}
	defer func() {
		if autoRsp {
			if resp.Code > 0 && resp.CodeMsg == "" {
				resp.CodeMsg = common.GetErrorMessage(resp.Code, resp.CodeMsg)
			}
			data = resp
		}
	}()

	request := &models.SendToUserMsgReq{}
	if err := json.Unmarshal(message, request); err != nil {
		resp.Code = common.ParameterIllegal
		zap.S().Info("SendToUserMsgReqController,  decode SendToUserMsgReq error, from: ", client.Addr, "seq: ", seq)
		return
	}

	zap.S().Info("SendToUserMsgReqController, receive SendToUserMsgReq", "addr", client.Addr, "ClientId", client.UniClientInfo.ClientId)

	msg := &models.UserMsgReq{
		AppId:   client.AppId,
		UserId:  client.UserId,
		Message: request.Message,
	}

	UnicastToClient(request.AppId, request.UserId, "usrmsg", msg)

	return
}

func UserMsgRspController(client *Client, seq string, message []byte) (autoRsp bool, data interface{}) {
	autoRsp = false

	request := &models.SendToUserMsgRsp{}
	if err := json.Unmarshal(message, request); err != nil {
		zap.S().Info("UserMsgRspController,  decode SendToUserMsgRsp error, from", client.Addr, "seq", seq)
		return
	}

	zap.S().Info("UserMsgRspController, receive SendToUserMsgRsp", "addr", client.Addr, "UserId", client.UserId, "from appid", request.AppId, "from userid", request.UserId)

	return
}
