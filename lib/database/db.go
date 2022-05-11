package database

import (
	"fmt"
	"strconv"
	"time"
	"gowebsocket/models"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

func DB() *gorm.DB {
	return db
}

// InitDB 初始化数据库
func InitDB() error {
	
	max_idle_conn, _ := strconv.Atoi(viper.GetString("mysql.max_idle_conn"))
	max_open_conn, _ := strconv.Atoi(viper.GetString("mysql.max_open_conn"))
	var err error
	db, err = openConn(viper.GetString("mysql.dsn"), max_idle_conn, max_open_conn)
	fmt.Println("InitDB ", viper.GetString("mysql.dsn"))
	if err != nil {
		return fmt.Errorf("open connection failed, error: %s", err.Error())
	}

	

	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&models.UserOnline{})
	

	return nil
}

func openConn(dsn string, idle, open int) (*gorm.DB, error) {
	newLogger := logger.New(Writer{}, logger.Config{
		SlowThreshold:             500 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false})
	openDB, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}
	db, err := openDB.DB()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(idle)
	db.SetMaxOpenConns(open)
	return openDB, nil
}

// Writer 记录SQL日志
type Writer struct{}

func (w Writer) Printf(format string, args ...interface{}) {
	zap.S().Debug(fmt.Sprintf(format, args...))
}
