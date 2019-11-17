package friends

import (
	"github.com/spkinger/speaker-server/src/module/auth"
	"github.com/spkinger/speaker-server/src/module/model"
	"github.com/spkinger/speaker-server/src/module/pub"
	"log"
	"net/http"
	"strconv"
)

// 添加好友申请
func AddRequest(writer http.ResponseWriter, request *http.Request) {
	// 己方uid
	uid := auth.GetUidRequest(request)
	if uid == 0 {
		pub.ErrRep(writer,"用户未登录", map[string]interface{}{})
		return
	}

	// 对方uid
	targetId := request.PostFormValue("target_id")
	targetId64, err := strconv.ParseInt(targetId, 10, 64)
	if err != nil {
		pub.ErrRep(writer,"用户不存在", map[string]interface{}{})
		return
	}

	if targetId64 == uid {
		pub.ErrRep(writer,"不能加自己为好友", map[string]interface{}{})
		return
	}

	target := model.User{}
	target.GetById(targetId64)
	if target.ID == 0 {
		pub.ErrRep(writer,"用户不存在.", map[string]interface{}{})
		return
	}

	// 添加好友申请
	friendRequest := model.FriendRequest{}
	if err := friendRequest.Add(uid, targetId64); err != nil {
		log.Print("add friend request err: ", err)
		pub.ErrRep(writer,"好友申请失败", map[string]interface{}{})
		return
	} else {
		pub.SuccRep(writer,"好友申请成功", map[string]interface{}{})
		return
	}
}

// 同意申请
func UpdateRequest(writer http.ResponseWriter, request *http.Request) {
	// 己方uid
	uid := auth.GetUidRequest(request)
	if uid == 0 {
		pub.ErrRep(writer,"用户未登录", map[string]interface{}{})
		return
	}

	// 对方uid
	targetId := request.PostFormValue("target_id")
	updateType := request.PostFormValue("type")
	targetId64, err := strconv.ParseInt(targetId, 10, 64)
	if err != nil {
		pub.ErrRep(writer,"好友申请记录不存在", map[string]interface{}{})
		return
	}

	switch updateType {
	case "1":
		// 同意申请
		friendRequest := model.FriendRequest{}
		if err := friendRequest.Apply(targetId64, uid); err != nil {
			log.Print("apply friend request err: ", err)
			pub.ErrRep(writer,"同意申请失败", map[string]interface{}{})
			return
		} else {
			pub.SuccRep(writer,"成功添加对方为好友", map[string]interface{}{})
			return
		}
	case "0":
		// 拒绝申请
		friendRequest := model.FriendRequest{}
		if err := friendRequest.Refuse(targetId64, uid); err != nil {
			log.Print("apply friend request err: ", err)
			pub.ErrRep(writer,"操作申请失败", map[string]interface{}{})
			return
		} else {
			pub.SuccRep(writer,"已拒绝对方申请", map[string]interface{}{})
			return
		}
	default:
		pub.ErrRep(writer,"请选择同意或拒绝申请", map[string]interface{}{})
		return
	}
}

// 删除好友
func DelFriends(writer http.ResponseWriter, request *http.Request) {
	uid := auth.GetUidRequest(request)
	if uid == 0 {
		pub.ErrRep(writer,"用户未登录", map[string]interface{}{})
		return
	}

	targetId := request.PostFormValue("target_id")
	targetId64, err := strconv.ParseInt(targetId, 10, 64)
	if err != nil {
		pub.ErrRep(writer,"用户不存在", map[string]interface{}{})
		return
	}

	target := model.User{}
	target.GetById(targetId64)
	if target.ID == 0 {
		pub.ErrRep(writer,"用户不存在", map[string]interface{}{})
		return
	}

	if Del(uid, targetId64) {
		pub.SuccRep(writer,"好友删除成功", map[string]interface{}{})
		return
	} else {
		pub.ErrRep(writer,"好友删除失败", map[string]interface{}{})
		return
	}
}

// 删除好友
func Del(from int64, target int64) bool {
	userRel := model.UserRel{}
	err := userRel.DelFriends(from, target)
	if err != nil {
		log.Print("delete friends err:", err)
		return false
	} else {
		return true
	}
}