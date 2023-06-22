/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-27
 * Time: 14:41
 */

package models

/************************  请求数据  **************************/
// 通用请求数据格式
type Request struct {
	Seq  string      `json:"seq"`            // 消息的唯一Id
	Cmd  string      `json:"cmd"`            // 请求命令字
	Data interface{} `json:"data,omitempty"` // 数据 json
}

// 登录请求数据
type LoginReq struct {
	AppId  string `json:"appId,omitempty"`
	UserId string `json:"userId,omitempty"`
	Name   string `json:"name,omitempty"`
}

// 心跳请求数据
// {"seq":"2324","cmd":"heartbeat","data":{"state"=0}}
type HeartBeatReq struct {
	State uint32 `json:"state"` //0=good&idle 1=busy
}

type SendToUserMsgReq struct {
	AppId   string `json:"appId,omitempty"`
	UserId  string `json:"userId,omitempty"`
	Message string `json:"message,omitempty"`
}

type UserMsgReq struct {
	AppId   string `json:"appId,omitempty"`
	UserId  string `json:"userId,omitempty"`
	Message string `json:"message,omitempty"`
}

type Ping struct {
	Ts int64 `json:"ts"`
}

type Credentials struct {
	Account  string `json:"account"`
	Password string `json:"password"` //md5(password)
}
