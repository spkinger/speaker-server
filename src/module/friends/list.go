package friends

import (
	"github.com/spkinger/speaker-server/src/module/auth"
	"github.com/spkinger/speaker-server/src/module/model"
	"github.com/spkinger/speaker-server/src/module/pub"
	"log"
	"net/http"
	"strconv"
)

// 好友申请列表--发出的
func RequestListSend(writer http.ResponseWriter, request *http.Request)  {
	// 己方uid
	uid := auth.GetUidRequest(request)
	if uid == 0 {
		return
	}

	// 页码
	pageInt, pageSizeInt := pub.GetPageParam(writer, request)

	// 查询
	friendRequest := model.FriendRequest{}
	requestList, err := friendRequest.FromList(uid, pageInt, pageSizeInt)
	if err != nil {
		log.Print("send request list err:", err)
		pub.ErrRep(writer, "查询失败", nil)
		return
	}

	pub.SuccRep(writer, "查询成功", requestList)
}

// 好友申请列表--收到的
func RequestListGot(writer http.ResponseWriter, request *http.Request) {
	// 己方uid
	uid := auth.GetUidRequest(request)
	if uid == 0 {
		return
	}

	// 页码
	pageInt, pageSizeInt := pub.GetPageParam(writer, request)

	// 查询
	friendRequest := model.FriendRequest{}
	requestList, err := friendRequest.TargetList(uid, pageInt, pageSizeInt)
	if err != nil {
		log.Print("got request list err:", err)
		pub.ErrRep(writer, "查询失败", nil)
		return
	}

	pub.SuccRep(writer, "查询成功", requestList)
}

// 我的好友
func MyFriends(writer http.ResponseWriter, request *http.Request) {
	// 己方uid
	uid := auth.GetUidRequest(request)
	if uid == 0 {
		return
	}

	// 页码
	pageInt, pageSizeInt := pub.GetPageParam(writer, request)

	// 查询
	userRel := model.UserRel{}
	myFriends, err := userRel.MyFriends(uid, pageInt, pageSizeInt)
	if err != nil {
		log.Print("got request list err:", err)
		pub.ErrRep(writer, "查询失败", nil)
		return
	}

	pub.SuccRep(writer, "查询成功", myFriends)
}

// 查找用户
func SearchUser(writer http.ResponseWriter, request *http.Request)  {
	nickName, ok := pub.GetParam(request, "nick_name")
	if !ok {
		pub.ErrRep(writer, "昵称不可为空", nil)
		return
	}

	// 页码
	pageInt, pageSizeInt := pub.GetPageParam(writer, request)

	// 查询
	user := model.User{}
	userList, err := user.SearchByNickName(nickName, pageInt, pageSizeInt)
	if err != nil {
		log.Print("got request list err:", err)
		pub.ErrRep(writer, "查询失败", nil)
		return
	}

	pub.SuccRep(writer, "查询成功", userList)
}

// 用户详情
func UserDetail(writer http.ResponseWriter, request *http.Request)  {
	userIdStr, ok := pub.GetParam(request, "user_id")
	if !ok {
		pub.ErrRep(writer, "用户ID不可为空", nil)
		return
	}

	// 查询
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		log.Print("userId to int64 err:", err)
		pub.ErrRep(writer, "用户ID不合法", nil)
		return
	}

	user := model.User{}
	userRes, err := user.Detail(userId)
	if err != nil {
		log.Print("got user detail err:", err)
		pub.ErrRep(writer, "查询失败", nil)
		return
	}

	pub.SuccRep(writer, "查询成功", userRes)
}