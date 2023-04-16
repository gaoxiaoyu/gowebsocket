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
type Login struct {
	ServiceToken string `json:"serviceToken"` // 验证用户是否登录
	AppId        uint32 `json:"appId,omitempty"`
	UserId       string `json:"userId,omitempty"`
}

// 心跳请求数据
// {"seq":"2324","cmd":"heartbeat","data":{"state"=0}}
type HeartBeat struct {
	State uint32 `json:"state"` //0=good&idle 1=bad 2=busy
}

// 登录请求数据
// 例子： {"seq":"2323","cmd":"register","data":{"uuid":"","state":0,"name":"xxx","group":0}}
type RegisterReq struct {
	Uuid  string `json:"uuid"`  //云手机的uuid
	State uint32 `json:"state"` //0=good 1=bad 2=busy
	Name  string `json:"name,omitempty"`
	Group uint32 `json:"group,omitempty"`
}

// 心跳请求数据
// 例子： {"seq":"2324","cmd":"cloudmobileheartbeat","data":{"state"=0}}
type CloudMobileHeartBeat struct {
	State uint32 `json:"state"` //0=good 1=bad 2=busy
}
