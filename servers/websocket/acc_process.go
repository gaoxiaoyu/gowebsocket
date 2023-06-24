/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-27
 * Time: 14:38
 */

package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"gowebsocket/auth"
	"gowebsocket/common"
	"gowebsocket/helper"
	"gowebsocket/models"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// type DisposeFunc func(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{})
type DisposeFunc func(client *Client, seq string, message []byte) (autoRsp bool, data interface{})

var (
	handlers        = make(map[string]DisposeFunc)
	handlersRWMutex sync.RWMutex
)

// 注册ws连接成功之后的json消息处理函数
func Register(key string, value DisposeFunc) {
	handlersRWMutex.Lock()
	defer handlersRWMutex.Unlock()
	handlers[key] = value

	return
}

func getHandlers(key string) (value DisposeFunc, ok bool) {
	handlersRWMutex.RLock()
	defer handlersRWMutex.RUnlock()

	value, ok = handlers[key]

	return
}

// 处理数据
func ProcessData(client *Client, message []byte) {

	fmt.Println("处理数据", client.Addr, string(message))

	defer func() {
		if r := recover(); r != nil {
			zap.S().Errorw("ProcessData, panic", "recover", r)
			helper.PrintStack(r)
		}
	}()

	request := &models.UniMessage{}

	err := json.Unmarshal(message, request)
	if err != nil {
		fmt.Println("处理数据 json Unmarshal", err)
		client.SendMsg([]byte("数据不合法"))
		return
	}
	zap.S().Infow("ProcessData from client", "addr", client.Addr, "command", request.Head.Cmd)

	if client.VerifyTime == 0 {
		if request.ClientInfo == nil || request.ClientInfo.AppId == "" {
			client.SendMsg([]byte("鉴权信息不全"))
			return
		}
		if err := VerifyClient(client, request.ClientInfo); err != nil {
			client.SendMsg([]byte("鉴权失败"))
			return
		}
		client.VerifyTime = uint64(time.Now().Unix())
		client.UniClientInfo = *request.ClientInfo
	}

	seq := request.Head.Seq
	cmd := request.Head.Cmd
	version := request.Head.Version

	var (
		autoRsp bool = true
		code    uint32
		data    interface{}
	)

	// request
	fmt.Println("acc_request", cmd, client.Addr)

	// 采用 map 注册的方式
	if value, ok := getHandlers(cmd); ok {
		autoRsp, data = value(client, seq, request.Data)
	} else {
		code = common.RoutingNotExist
		fmt.Println("处理数据 路由不存在", client.Addr, "cmd", cmd)
	}

	if autoRsp {
		var responseMsg *models.UniMessage
		if code > 0 {
			responseMsg = models.PrepareUniMessageWithCode(seq, cmd, version, code, "", data)

		} else {
			responseMsg = models.PrepareUniMessage(seq, cmd, version, data)
		}

		response, err := json.Marshal(responseMsg)
		if err != nil {
			fmt.Println("处理数据 json Marshal", err)
			return
		}

		client.SendMsg(response)

		fmt.Println("acc_response send", client.Addr, client.AppId, client.UserId, "cmd", cmd, "code", code, "data", data)
	}

	return
}

func VerifyClient(client *Client, clientinfo *models.UniClientInfo) error {
	//todo@: 未来集成jwt，验证长连接接入信息,并且验证appid
	myclaim, err := auth.JwtVerify(clientinfo.ClientToken)
	if err != nil {
		zap.S().Errorw("VerifyClient, JwtVerify err", "err", err, "clientid", clientinfo.ClientId)
		return err
	}

	if clientinfo.ClientId != strconv.FormatUint(myclaim.UserId, 10) {
		zap.S().Errorw("VerifyClient, clientid mismatch", "userid", myclaim.UserId, "clientid", clientinfo.ClientId)
		return errors.New("jwt userid not equal to clientid")
	}
	zap.S().Debugw("VerifyClient, clientid verify pass", "userid", myclaim.UserId, "IssuedAt", myclaim.IssuedAt)

	return nil
}
