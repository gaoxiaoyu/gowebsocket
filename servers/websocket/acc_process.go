/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-27
 * Time: 14:38
 */

package websocket

import (
	"encoding/json"
	"fmt"
	"gowebsocket/common"
	"gowebsocket/models"
	"sync"
	"go.uber.org/zap"
)

type DisposeFunc func(client *Client, seq string, message []byte) (autoRsp bool, code uint32, msg string, data interface{})

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
			fmt.Println("处理数据 stop", r)
		}
	}()

	request := &models.Request{}

	err := json.Unmarshal(message, request)
	if err != nil {
		fmt.Println("处理数据 json Unmarshal", err)
		client.SendMsg([]byte("数据不合法"))

		return
	}
	zap.S().Info("ProcessData from client: ", client.Addr, " command:", request.Cmd)

	requestData, err := json.Marshal(request.Data)
	if err != nil {
		fmt.Println("处理数据 json Marshal", err)
		client.SendMsg([]byte("处理数据失败"))

		return
	}

	seq := request.Seq
	cmd := request.Cmd

	var (
		autoRsp bool = true
		code    uint32
		msg     string
		data    interface{}
	)

	// request
	fmt.Println("acc_request", cmd, client.Addr)

	// 采用 map 注册的方式
	if value, ok := getHandlers(cmd); ok {
		autoRsp, code, msg, data = value(client, seq, requestData)
	} else {
		code = common.RoutingNotExist
		fmt.Println("处理数据 路由不存在", client.Addr, "cmd", cmd)
	}

	if autoRsp {

		msg = common.GetErrorMessage(code, msg)

		responseHead := models.NewResponseHead(seq, cmd, code, msg, data)

		headByte, err := json.Marshal(responseHead)
		if err != nil {
			fmt.Println("处理数据 json Marshal", err)

			return
		}

		client.SendMsg(headByte)

		fmt.Println("acc_response send", client.Addr, client.AppId, client.UserId, "cmd", cmd, "code", code, "data", data)
	}

	return
}
