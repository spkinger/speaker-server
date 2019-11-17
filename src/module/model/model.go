package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

var Db *gorm.DB

// 初始化数据库连接
func CreateDb()  {
	var err error
	Db, err = gorm.Open("mysql", "root:123456@/speaker?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local")
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