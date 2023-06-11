package models

import "gorm.io/gorm"

// 用户在线状态
type UserOnlineInDb struct {
	gorm.Model
	AppId         string `gorm:"index:idx_uid, unique, priority:9; size:191"`  // appId
	UserId        string `gorm:"index:idx_uid, unique, priority:10; size:191"` // 用户Id
	ClientType    uint32 // clientType
	ClientId      string // 用户Id
	Name          string // name
	Platform      string // 平台
	Ua            string // ua
	ClientIp      string // 客户端Ip
	ClientPort    string // 客户端端口
	LoginTime     uint64 // 用户上次登录时间
	HeartbeatTime uint64 // 用户上次心跳时间
	LogOutTime    uint64 // 用户退出登录的时间
	AccIp         string // acc Ip
	AccPort       string // acc 端口
}

func (UserOnlineInDb) TableName() string {
	return "UserOnline"
}
