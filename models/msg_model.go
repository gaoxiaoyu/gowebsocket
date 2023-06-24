/**
* Created by GoLand.
* User: link1st
* Date: 2019-08-01
* Time: 10:40
 */

package models

import (
	"encoding/json"
	"gowebsocket/common"
)

const (
	MessageTypeText = "text"
	MessageCmdMsg   = "msg"
	MessageCmdEnter = "enter"
	MessageCmdExit  = "exit"
)

const (
	UniMsgVersion1Define uint32 = 1
)

// 消息的定义
type Message struct {
	Target string `json:"target"` // 目标
	Type   string `json:"type"`   // 消息类型 text/img/
	Msg    string `json:"msg"`    // 消息内容
	From   string `json:"from"`   // 发送者
}

func NewTestMsg(from string, Msg string) (message *Message) {

	message = &Message{
		Type: MessageTypeText,
		From: from,
		Msg:  Msg,
	}

	return
}

func getTextMsgData(cmd, uuId, msgId, message string) string {
	textMsg := NewTestMsg(uuId, message)
	head := NewResponseHead(msgId, cmd, common.OK, "Ok", textMsg)

	return head.String()
}

// 文本消息
func GetMsgData(uuId, msgId, cmd, message string) string {

	return getTextMsgData(cmd, uuId, msgId, message)
}

// 文本消息
func GetTextMsgData(uuId, msgId, message string) string {

	return getTextMsgData("msg", uuId, msgId, message)
}

// 用户进入消息
func GetTextMsgDataEnter(uuId, msgId, message string) string {

	return getTextMsgData("enter", uuId, msgId, message)
}

// 用户退出消息
func GetTextMsgDataExit(uuId, msgId, message string) string {

	return getTextMsgData("exit", uuId, msgId, message)
}

type UniHead struct {
	Seq     string `json:"seq"`     // 消息的唯一Id
	Cmd     string `json:"cmd"`     // 请求命令字
	Version uint32 `json:"version"` //协议版本
}

type UniClientInfo struct {
	AppId       string `json:"appId,omitempty"`
	ClientType  uint32 `json:"clientType,omitempty"`
	ClientId    string `jsons:"clientId,omitempty"`
	ClientToken string `json:"clientToken,omitempty"`
	Platform    string `json:"platform,omitempty"` //平台类型：“android” “ios” "pc" "web" + & +平台版本:“1.0.0”
	Ua          string `json:"ua,omitempty"`       //“appName&appVersion&渠道”
}

type RspCodeInfo struct {
	Code    uint32 `json:"code,omitempty"`
	CodeMsg string `json:"codeMsg,omitempty"`
}

func (r *RspCodeInfo) IsEmpty() bool {
	return r.Code == 0 && r.CodeMsg == ""
}

type UniMessage struct {
	Head       UniHead         `json:"head,omitempty"`
	ClientInfo *UniClientInfo  `json:"clientinfo,omitempty"`
	RspCode    *RspCodeInfo    `json:"rspCode,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"` // 数据 json
}

type UnicastMessage struct {
	AppId  string
	UserId string
	Data   []byte
}

func PrepareUniMessage(seq string, cmd string, version uint32, data interface{}) *UniMessage {
	jsondata, _ := json.Marshal(data)
	unimsg := &UniMessage{
		Head: UniHead{Seq: seq, Cmd: cmd, Version: version},
		Data: jsondata,
	}
	return unimsg
}

func BuildUniMessage(seq string, cmd string, version uint32, data interface{}) []byte {
	jsondata, _ := json.Marshal(data)
	unimsg := &UniMessage{
		Head: UniHead{Seq: seq, Cmd: cmd, Version: version},
		Data: jsondata,
	}

	jsonUnimsg, _ := json.Marshal(unimsg)

	return jsonUnimsg
}

func PrepareUniMessageWithCode(seq string, cmd string, version, code uint32, codemsg string, data interface{}) *UniMessage {
	jsondata, _ := json.Marshal(data)
	if codemsg == "" {
		codemsg = common.GetErrorMessage(code, codemsg)
	}
	unimsg := &UniMessage{
		Head:    UniHead{Seq: seq, Cmd: cmd, Version: version},
		RspCode: &RspCodeInfo{Code: code, CodeMsg: codemsg},
		Data:    jsondata,
	}
	return unimsg
}
