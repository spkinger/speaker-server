package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spkinger/speaker-server/src/module/config"
	"log"
)

var Db *gorm.DB

// 初始化数据库连接
func CreateDb()  {
	var err error
	connInfo := fmt.Sprintf("%s:%s@/%s", config.Config.DBUser, config.Config.DBPassword, config.Config.DBName)
	Db, err = gorm.Open("mysql", connInfo+"?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("mysql connect err: ", err)
	}
}

// 关闭数据库连接
func CloseDb()  {
	if Db == nil {
		return
	}
	err := Db.Close()
	if err != nil {
		log.Println("mysql close err: ", err)
	}
}

// 初始化数据库
func InitTable()  {
	Db.AutoMigrate(&Admin{}, &User{}, &UserRel{}, &FriendRequest{})
	log.Println("tables created！")
}