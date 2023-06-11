package helper

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

var node *snowflake.Node

func init() {
	var st time.Time
	var currentTime string = time.Now().Format("2006-01-02")
	st, err := time.Parse("2006-01-02", currentTime)
	if err != nil {
		zap.S().DPanic("id generator init, time.Parse err", "err", err.Error())
		return
	}
	snowflake.Epoch = st.UnixNano() / 1000000
	if node, err = snowflake.NewNode(1); err != nil {
		zap.S().DPanic("id generator init, NewNode err", "err", err.Error())
	}
}

func GenUint64Id() (id uint64) {
	id = uint64(node.Generate().Int64())
	zap.S().Debug("GenUint64Id", "id")
	return
}
