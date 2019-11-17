package model

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type User struct {
	ID int64
	Name string `sql:"comment:'账户'"`
	NickName string `sql:"comment:'昵称'"`
	Passwd string `sql:"comment:'密码'"`
	Status int8 `gorm:"default:1"sql:"comment:'用户状态：1正常，0暂停，-1禁用'"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type UserResult struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	NickName string `json:"nick_name"`
}

// 创建用户
func (u *User) Create() error {
	return Db.Create(u).Error
}

// 获取用户-通过账户
func (u *User) GetByName(name string) {
	Db.Where("name = ?", name).First(u)
}

// 获取用户-昵称
func (u *User) GetByNickName(nickName string) {
	Db.Where("nick_name = ?", nickName).First(u)
}

// 获取用户-ID
func (u *User) GetById(id int64) {
	Db.Where("id = ?", id).First(u)
}

// 用户详情--对外开放
func (u *User) Detail(id int64) (*UserResult, error) {
	u.GetById(id)
	if u == nil {
		return nil, errors.New("用户不存在")
	}

	userResult := UserResult{
		ID: u.ID,
		Name: u.Name,
		NickName: u.NickName,
	}
	return &userResult, nil
}

// 根据昵称查找用户--对外开放
func (u *User) SearchByNickName(nickName string, page int, pageSize int) ([]UserResult, error) {
	userResultList := []UserResult{}
	offset := (page - 1) * pageSize
	rows, err := Db.Table("users").
		Where("nick_name LIKE ? and status = 1", fmt.Sprintf("%%%s%%", nickName)).
		Select("id, name, nick_name").
		Order("nick_name asc").
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