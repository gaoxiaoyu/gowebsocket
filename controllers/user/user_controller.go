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
