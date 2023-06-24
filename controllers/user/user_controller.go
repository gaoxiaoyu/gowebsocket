/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-25
* Time: 12:11
 */

package user

import (
	"fmt"
	"gowebsocket/auth"
	"gowebsocket/common"
	"gowebsocket/controllers"
	"gowebsocket/helper"
	"gowebsocket/lib/cache"
	"gowebsocket/lib/database"
	"gowebsocket/models"
	"gowebsocket/servers/websocket"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	//appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	data := make(map[string]interface{})

	online := websocket.CheckUserOnline(appIdStr, userId)
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

	//appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	data := make(map[string]interface{})

	if cache.SeqDuplicates(msgId) {
		fmt.Println("给用户发送消息 重复提交:", msgId)
		controllers.Response(c, common.OK, "", data)

		return
	}

	sendResults, err := websocket.SendUserMessage(appIdStr, userId, msgId, message)
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

	//appId, _ := strconv.ParseInt(appIdStr, 10, 32)

	data := make(map[string]interface{})
	if cache.SeqDuplicates(msgId) {
		fmt.Println("给用户发送消息 重复提交:", msgId)
		controllers.Response(c, common.OK, "", data)

		return
	}

	sendResults, err := websocket.SendUserMessageAll(appIdStr, userId, msgId, models.MessageCmdMsg, message)
	if err != nil {
		data["sendResultsErr"] = err.Error()

	}

	data["sendResults"] = sendResults

	controllers.Response(c, common.OK, "", data)

}

func Register(c *gin.Context) {
	creds := models.Credentials{}

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	//check local user account mysql
	var user models.RegisterUser
	if err := database.DB().Debug().Where("account = ?", creds.Account).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}
	}

	if user.Account == creds.Account {
		c.AbortWithStatusJSON(
			http.StatusMethodNotAllowed,
			gin.H{"error": "账号已存在"})
		return
	}

	user.Account = creds.Account
	user.Password = creds.Password
	user.UserId = helper.GenUint64Id()

	if err := database.DB().Debug().Create(&user).Error; err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"userId": user.UserId})

}

func Signin(c *gin.Context) {
	creds := models.Credentials{}

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	var user models.RegisterUser
	if err := database.DB().Debug().Where("account = ?", creds.Account).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	if user.Password != creds.Password {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "密码错误"})
		return
	}

	expirationTime := time.Now().Add(180 * time.Minute)
	// 创建JWT声明，其中包括用户名和有效时间
	claims := &auth.Claims{
		UserId:  user.UserId,
		Account: creds.Account,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 创建JWT字符串
	tokenString, err := token.SignedString(auth.JwtKey)

	if err != nil {
		// 如果创建JWT时出错，则返回内部服务器错误
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString, "userId": user.UserId})

}

func JwtAuth(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "header缺少Authorization字段"})
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "请求头中auth格式有误"})
		return
	}

	claims := &auth.Claims{}
	tknStr := parts[1]
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return auth.JwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": err})
			return
		}

		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"error": err})
		return
	}

	if !tkn.Valid {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "token invalid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": claims.Account, "userId": claims.UserId})
}
