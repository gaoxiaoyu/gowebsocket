/**
* Created by GoLand.
* User: link1st
* Date: 2019-08-01
* Time: 10:46
 */

package models

import "encoding/json"

/************************  响应数据  **************************/
type Head struct {
	Seq      string    `json:"seq"`      // 消息的Id
	Cmd      string    `json:"cmd"`      // 消息的cmd 动作
	Response *Response `json:"response"` // 消息体
}

type Response struct {
	Code    uint32      `json:"code"`
	CodeMsg string      `json:"codeMsg"`
	Data    interface{} `json:"data"` // 数据 json
}

// push 数据结构体
type PushMsg struct {
	Seq  string `json:"seq"`
	Uuid uint64 `json:"uuid"`
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

type RegisterRsp struct {
	Rtc_channel uint16 `json:"rtc_channel"` 
	Signal_channel uint16 `json:"signal_channel"`
}

// {"seq":"2324","cmd":"heartbeat","data":{"state"=0}}
type HeartBeatRsp struct {
	State uint32 `json:"state,omitempty "` //0=good&idle 1=bad 2=busy
}

// AssignedReq，分配请求
// 例子：{"seq":"2323","cmd":"assign","response":{"code":200,"codeMsg":"Success","data":{"uid":uid,"rtc_channel": rtc_channel,"signal_channel":signal_channel}}}
type AssignedReq struct {
	Uid  uint32 `json:"uid"`  //
	Rtc_channel uint64 `json:"rtc_channel"` //
	Signal_channel uint64 `json:"signal_channel"`
}

type RecyleReq struct {
	Uid  uint32 `json:"uid"`  //
	Rtc_channel uint64 `json:"rtc_channel"` //
	Signal_channel uint64 `json:"signal_channel"`
}


// 设置返回消息
func NewResponseHead(seq string, cmd string, code uint32, codeMsg string, data interface{}) *Head {
	response := NewResponse(code, codeMsg, data)

	return &Head{Seq: seq, Cmd: cmd, Response: response}
}

func (h *Head) String() (headStr string) {
	headBytes, _ := json.Marshal(h)
	headStr = string(headBytes)

	return
}

func NewResponse(code uint32, codeMsg string, data interface{}) *Response {
	return &Response{Code: code, CodeMsg: codeMsg, Data: data}
}
