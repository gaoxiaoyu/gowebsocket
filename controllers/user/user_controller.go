/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-25
* Time: 12:11
 */

package user

import (
	"fmt"
	"gowebsocket/common"
	"gowebsocket/controllers"
	"gowebsocket/lib/cache"
	"gowebsocket/models"
	"gowebsocket/servers/websocket"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 查看全部在线用户
func List(c *gin.Context) {

	appIdStr := c.Query("appId")
	appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	fmt.Println("http_request 查看全部在线用户", appId)

	data := make(map[string]interface{})

	userList := websocket.UserList()
	data["userList"] = userList

	controllers.Response(c, common.OK, "", data)
}

// 查看用户是否在线
func Online(c *gin.Context) {

	userId := c.Query("userId")
	appIdStr := c.Query("appId")

	fmt.Println("http_request 查看用户是否在线", userId, appIdStr)
	appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	data := make(map[string]interface{})

	online := websocket.CheckUserOnline(uint32(appId), userId)
	data["userId"] = userId
	data["online"] = online

	controllers.Response(c, common.OK, "", data)
}

// 给用户发送消息
func SendMessage(c *gin.Context) {
	// 获取参数
	appIdStr := c.PostForm("appId")
	userId := c.PostForm("userId")
	msgId := c.PostForm("msgId")
	message := c.PostForm("message")

	fmt.Println("http_request 给用户发送消息", appIdStr, userId, msgId, message)

	appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	data := make(map[string]interface{})

	if cache.SeqDuplicates(msgId) {
		fmt.Println("给用户发送消息 重复提交:", msgId)
		controllers.Response(c, common.OK, "", data)

		return
	}

	sendResults, err := websocket.SendUserMessage(uint32(appId), userId, msgId, message)
	if err != nil {
		data["sendResultsErr"] = err.Error()
	}

	data["sendResults"] = sendResults

	controllers.Response(c, common.OK, "", data)
}

// 给全员发送消息
func SendMessageAll(c *gin.Context) {
	// 获取参数
	appIdStr := c.PostForm("appId")
	userId := c.PostForm("userId")
	msgId := c.PostForm("msgId")
	message := c.PostForm("message")

	fmt.Println("http_request 给全体用户发送消息", appIdStr, userId, msgId, message)

	appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	data := make(map[string]interface{})
	if cache.SeqDuplicates(msgId) {
		fmt.Println("给用户发送消息 重复提交:", msgId)
		controllers.Response(c, common.OK, "", data)

		return
	}

	sendResults, err := websocket.SendUserMessageAll(uint32(appId), userId, msgId, models.MessageCmdMsg, message)
	if err != nil {
		data["sendResultsErr"] = err.Error()

	}

	data["sendResults"] = sendResults

	controllers.Response(c, common.OK, "", data)

}

//func StartXRLive(c *gin.Context) {
//
//	userId := c.PostForm("userId")
//	//token := c.Query("token")
//	//todo@: verifytoken，等鹏爷给接口
//	fmt.Println("StartXRLive 请求XRLive from user：", userId)
//
//	ch := make(chan *websocket.AllocateInfo)
//	quit := make(chan bool)
//
//	data := make(map[string]interface{})
//
//	uid, _ := strconv.Atoi(userId)
//	result, _ := websocket.AllocateCloudMobile(uint32(uid), ch)
//	if !result {
//		data := make(map[string]interface{})
//		data["userId"] = userId
//		controllers.Response(c, common.NoResource, "资源不足", data)
//		return
//	}
//	var allocateResult *websocket.AllocateInfo
//
//	select {
//	case allocateResult = <-ch:
//		fmt.Println("Got allocate result: ", allocateResult)
//
//	case <-time.After(3 * time.Second):
//		fmt.Println("StartXRLive TimeOut")
//		quit <- true
//	}
//	<-quit
//
//	data["userId"] = userId
//	data["code"] = allocateResult.Code
//	data["rtc_channel"] = allocateResult.Rtc_channel
//	data["signal_channel"] = allocateResult.Signal_channel
//	data["cloudmobile_uuid"] = allocateResult.Uuid
//	data["cloudmobile_group"] = allocateResult.Group
//
//	controllers.Response(c, common.OK, "", data)
//}
// how to test:
// curl http://192.168.2.9:8080/user/StartXRLive?userId=123
func StartXRLive(c *gin.Context) {

	userId := c.Query("userId")
	//token := c.Query("token")
	//todo@: verifytoken，等鹏爷给接口
	fmt.Println("StartXRLive 请求XRLive from user：", userId)

	data := make(map[string]interface{})

	uid, _ := strconv.Atoi(userId)
	result, rtcChannel, signalChannel := websocket.AllocateCloudMobile(uint32(uid))
	fmt.Println("StartXRLive 云手机分配结果", userId, result, uint16(rtcChannel), uint16(signalChannel))

	if !result {
		data := make(map[string]interface{})
		data["userId"] = userId
		controllers.Response(c, common.NoResource, "资源不足", data)
		return
	}

	data["userId"] = userId
	data["rtcChannel"] = uint16(rtcChannel)
	data["signalChannel"] = uint16(signalChannel)

	controllers.Response(c, common.OK, "", data)
}

// curl http://192.168.2.9:8080/user/StopXRLive?userId=123
func StopXRLive(c *gin.Context) {

	userId := c.Query("userId")
	//token := c.Query("token")
	//todo@: verifytoken，等鹏爷给接口
	fmt.Println("StopXRLive 停止XRLive from user：", userId)

	data := make(map[string]interface{})
	data["userId"] = userId

	uid, _ := strconv.Atoi(userId)
	result := websocket.RecyleCloudMobile(uint32(uid))
	fmt.Println("StopXRLive 云手机回收结果", userId, result)


	if !result {
		controllers.Response(c, common.NoResource, "资源不足", data)
		return
	}


	controllers.Response(c, common.OK, "", data)
}
