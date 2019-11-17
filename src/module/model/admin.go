package model

import "time"

type Admin struct {
	ID int64
	Name string `sql:"comment:'管理员账户'"`
	Passwd string `sql:"comment:'密码'"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}