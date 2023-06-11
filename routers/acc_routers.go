/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 16:02
 */

package routers

import (
	"gowebsocket/servers/websocket"
)

// Websocket 路由
// ws消息处理函数都在这里注册
func WebsocketInit() {
	websocket.Register("login", websocket.LoginController)
	websocket.Register("heartbeat", websocket.HeartbeatController)
	websocket.Register("sendtousrmsg", websocket.SendToUserMsgReqController)
	websocket.Register("usrmsg", websocket.UserMsgRspController)
}
