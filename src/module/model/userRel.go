package model

import (
	"github.com/jinzhu/gorm"
	"log"
)

// 好友关系表
type UserRel struct {
	ID int64
	FromUser int64 `sql:"comment:'用户ID'"`
	TargetUser int64 `sql:"comment:'用户ID'"`
}

// 添加好友
func (ur *UserRel) AddFriends(tx *gorm.DB,from int64, target int64) error {
	urFrom := UserRel{
		FromUser:from,
		TargetUser:target,
	}
	urTarget := UserRel{
		FromUser:target,
		TargetUser:from,
	}

	if err := tx.Create(&urFrom).Error; err != nil {
		return err
	}

	if err := tx.Create(&urTarget).Error; err != nil {
		return err
	}

	return nil
}

// 删除好友
func (ur *UserRel) DelFriends(from int64, target int64) error {
	tx := Db.Begin()

	if err := tx.Where("from_user = ? AND target_user = ?", from, target).Delete(UserRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("from_user = ? AND target_user = ?", target, from).Delete(UserRel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// 我的好友列表
func (ur *UserRel) MyFriends(from int64, page int, pageSize int) ([]UserResult, error) {
	userResultList := []UserResult{}
	offset := (page - 1) * pageSize
	//Db.LogMode(true)
	rows, err := Db.Table("user_rels").Joins("JOIN users ON users.id = user_rels.target_user").
				Where("user_rels.from_user = ?", from).
				Select("users.id, users.name, users.nick_name").
				Order("users.nick_name asc").
				Offset(offset).
				Limit(pageSize).
				Rows()

	if err != nil {
		return userResultList, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Print("rows close err:", err)
		}
	}()

	for rows.Next() {
		var userResult UserResult
		err := Db.ScanRows(rows, &userResult)
		if err != nil {
			return userResultList, err
		}
		userResultList = append(userResultList, userResult)
	}

	return userResultList, nil
}