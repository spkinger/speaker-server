package model

import (
	"errors"
	"log"
	"time"
)

const FriendRequestStatusStart = 0
const FriendRequestStatusApply = 1
const FriendRequestStatusRefuse = 2

// 好友申请表
type FriendRequest struct {
	ID int64
	FromUser int64 `sql:"comment:'用户ID'"`
	TargetUser int64 `sql:"comment:'用户ID'"`
	Status int8 `sql:"comment:'申请状态：0开启申请，1同意申请，2拒绝申请'"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// 添加申请
func (fr *FriendRequest)Add(from int64, target int64) error {
	fr.GetRequestByFT(from, target)
	if fr.ID == 0 {
		// 新建记录
		fr.FromUser = from
		fr.TargetUser = target
		fr.Status = FriendRequestStatusStart
		return Db.Create(fr).Error
	} else {
		// 更新记录
		fr.Status = FriendRequestStatusStart
		return Db.Save(fr).Error
	}
}

// 同意申请
func (fr *FriendRequest)Apply(fromId int64, targetId int64) error {
	// 对方的申请
	fr.GetRequestByFT(fromId, targetId)
	if fr.ID == 0 {
		return errors.New("记录不存在")
	}
	if fr.TargetUser != targetId {
		return errors.New("不能操作他人申请")
	}

	// 己方的申请（可能不存在）
	myFr := FriendRequest{}
	myFr.GetRequestByFT(fr.TargetUser, fr.FromUser)

	// 设置对方的申请为同意
	tx := Db.Begin()
	fr.Status = 1
	if err := tx.Save(fr).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 将己方的申请同样设置为同意
	if myFr.ID != 0 {
		myFr.Status = 1
		if err := tx.Save(&myFr).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 添加用户好友关系
	userRel := UserRel{}
	err := userRel.AddFriends(tx, fr.FromUser, fr.TargetUser)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// 拒绝申请
func (fr *FriendRequest)Refuse(fromId int64, targetId int64) error {
	fr.GetRequestByFT(fromId, targetId)
	if fr.ID == 0 {
		return errors.New("记录不存在")
	}

	if fr.TargetUser != targetId {
		return errors.New("不能操作他人申请")
	}

	fr.Status = 2
	return Db.Save(fr).Error
}

// 查找一条申请记录
func (fr *FriendRequest) GetRequestByFT(from int64, target int64) {
	Db.Where("from_user = ? and target_user = ?", from, target).First(fr)
}

func (fr *FriendRequest) GetRequestById(id int64) {
	Db.First(fr, id)
}

// 好友申请列表元素
type RequestUser struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	NickName string `json:"nick_name"`
	Status int8 `json:"status"`
}

// 好友申请列表-我发出的
func (fr *FriendRequest)FromList(fromUid int64, page int, pageSize int) ([]RequestUser, error) {
	offset := (page - 1) * pageSize
	requestUserList := []RequestUser{}
	rows, err :=Db.Table("friend_requests").Joins("JOIN users On users.id = friend_requests.target_user").
				Where("friend_requests.from_user = ?", fromUid).
				Select("users.id, users.name, users.nick_name, friend_requests.status").
				Order("friend_requests.updated_at desc").
				Offset(offset).
				Limit(pageSize).
				Rows()
	if err != nil {
		return requestUserList, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Print("rows close err:", err)
		}
	}()

	for rows.Next() {
		var requestUser RequestUser
		err := Db.ScanRows(rows, &requestUser)
		if err != nil {
			return requestUserList, err
		}
		requestUserList = append(requestUserList, requestUser)
	}

	return requestUserList, nil
}

// 好友申请列表-我收到的
func (fr *FriendRequest)TargetList(targetUid int64, page int, pageSize int) ([]RequestUser, error) {
	offset := (page - 1) * pageSize
	requestUserList := []RequestUser{}
	rows, err :=Db.Table("friend_requests").Joins("JOIN users ON users.id = friend_requests.from_user").
				Where("friend_requests.target_user = ?", targetUid).
				Select("users.id, users.name, users.nick_name, friend_requests.status").
				Order("friend_requests.updated_at desc").
				Offset(offset).
				Limit(pageSize).
				Rows()
	if err != nil {
		return requestUserList, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Print("rows close err:", err)
		}
	}()

	for rows.Next() {
		var requestUser RequestUser
		err := Db.ScanRows(rows, &requestUser)
		if err != nil {
			return requestUserList, err
		}
		requestUserList = append(requestUserList, requestUser)
	}

	return requestUserList, nil
}