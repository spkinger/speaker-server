package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/spkinger/speaker-server/src/module/auth"
	"github.com/spkinger/speaker-server/src/module/config"
	"github.com/spkinger/speaker-server/src/module/friends"
	"github.com/spkinger/speaker-server/src/module/middleware"
	"github.com/spkinger/speaker-server/src/module/model"
	"github.com/spkinger/speaker-server/src/module/pub"
	"github.com/spkinger/speaker-server/src/module/route"
	"log"
	"net/http"
	"os"
)

// websocket配置
var ws = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// 用户管道
var userChanList = map[string]chan string{}

// 获取config文件路径
var configPath = flag.String("C", "./config.json", "config file path")


func init() {
	log.SetPrefix("【Speaker-Server】")
	log.SetFlags(log.Ldate|log.Ltime |log.LUTC)

	// 读取配置文件
	config.ReadConfig(*configPath)

	// 判断允许的域名
	ws.CheckOrigin = func(r *http.Request) bool {
		log.Println("request host: "+r.Header.Get("Origin"))
		return pub.StrInSlice(r.Header.Get("Origin"), config.Config.WssAllowOrigin)
	}

	// 连接数据库
	model.CreateDb()
}

func main() {
	// 项目初始化
	if len(os.Args) >= 2 && os.Args[1] == "init" {
		initProject()
		return
	}

	// 注册退出前执行的事件
	defer beforeExit()

	//tst()
	//return

	// 设置路由
	router := route.InitRouter()
	router.Use(middleware.UserStatusCheck)
	router.AddFuc("/wss", wsHandler) // http升级为wss
	router.AddFuc("/register", auth.Register) // 注册接口
	router.AddFuc("/login", auth.Login) // 登录接口
	router.AddFuc("/friend/request/add", friends.AddRequest) // 添加好友申请
	router.AddFuc("/friend/request/update", friends.UpdateRequest) // 同意\拒绝好友申请
	router.AddFuc("/friend/del", friends.DelFriends) // 好友删除
	router.AddFuc("/friend/request/my/send", friends.RequestListSend) // 我发出的好友申请
	router.AddFuc("/friend/request/my/got", friends.RequestListGot) // 我收到的好友申请
	router.AddFuc("/friend/my", friends.MyFriends) // 我的好友
	router.AddFuc("/user/search", friends.SearchUser) // 用户搜索
	router.AddFuc("/user/detail", friends.UserDetail) // 用户详情
	router.AddFuc("/test", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("test in")
		_, err := writer.Write([]byte("test"))
		if err != nil {
			log.Println("tst: ", err)
		}
	})

	// 启动https服务
	router.ServeTLS(
		config.Config.HttpAddr,
		config.Config.TSLCertFile,
		config.Config.TSLKeyFile)
	//router.ServeHttp("0.0.0.0:80")
}

func tst()  {
	//user := model.User{}
	//user.GetByName("aaa")
	//log.Print(user)
	//ur := model.UserRel{
	//	FromUser: int64(1),
	//	TargetUser: int64(2),
	//}
	//res := ur.Create()
	//log.Printf("create: %v", res)
	//ur.Delete(1, 2)
	//fr := model.FriendRequest{}
	//err := fr.Add(1, 2)
	//log.Print("fr:", fr)
	//log.Print("err:", err)
}

// 初始化项目
func initProject()  {
	// 建表
	model.InitTable()
}

// 退出前处理
func beforeExit()  {
	// 异常捕获，保证正常退出
	log.Println("before exit")
	if err := recover(); err != nil {
		log.Println("recover success:", err)
	}

	// 关闭数据库连接
	model.CloseDb()
}

// http升级为websocket
func wsHandler(writer http.ResponseWriter, request *http.Request) {
	userID := int64(0)
	conn, err := ws.Upgrade(writer, request, nil)
	if err != nil {
		log.Println("websocket update: ", err)
		return
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("websocket close:", err)
		}
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("websocket read msg: ", err)
			break
		}
		log.Println("websocket receive msg: ", string(p))

		if messageType == websocket.TextMessage {
			if userID == int64(0) {
				userID = wsOnMessage(conn, p)
			} else {
				wsOnMessage(conn, p)
			}
		} else {
			log.Println("messageType：not TextMessage")
		}
	}

	// socket关闭，通知管道关闭
	chanName := getChanName(userID)
	if _,ok := userChanList[chanName]; ok {
		userChanList[chanName] <- "close"
		close(userChanList[chanName])
	}
}

// websocket消息处理
type ClientMsg struct {
	Auth    auth.AuthUser `json:"auth"` // 权限信息
	MsgType int `json:"type"` // 消息类型
	ChildType int `json:"child_type"` // 消息子类型
	Target  int64 `json:"target"` // 发送目标
	Data    interface{} `json:"data"` // 数据
}

const MsgTypeSys = 1 // 客户端与服务端通信
const MsgTypeUser = 2 // 客户端与客户端的通信
const ChildTypeSysInit = 1 // 初始化用户管道
const ChildTypeRefreshToken = 2 // 刷新token
const ChildTypeUserICE = 1 // ICE
const ChildTypeUserOffer = 2 // offer
const ChildTypeUserAnswer = 3 // answer
const ChildTypeUserPing = 4 // ping通信对方是否在线
const ChildTypeUserPong = 5 // 响应ping的一端
const ChildTypeUserCall = 6 // 呼叫对方
const ChildTypeUserCallAccess = 7 // 接受呼叫
const ChildTypeUserCallRefuse = 8 // 拒绝呼叫
const ChildTypeUserHangUp = 9 // 挂机

type ChanMsg struct {
	MsgType int `json:"type"`
	ChildType int `json:"child_type"` // 消息子类型
	From  int64 `json:"from"`
	Data  interface{} `json:"data"`
}

// 收到消息的处理
func wsOnMessage(conn *websocket.Conn, msg []byte) int64 {
	var clientMsg ClientMsg
	err := json.Unmarshal(msg, &clientMsg)

	if err != nil {
		log.Println("ws msg unmarshal: ", err)
		return 0
	}

	// 验证用户身份
	checkRes, newAuthUser := auth.CheckUser(clientMsg.Auth)
	if !checkRes {
		log.Println("ws user check err: ", clientMsg.Auth)
		return 0
	}

	userID       := clientMsg.Auth.ID
	userChanName := getChanName(userID)

	// 用户token过期通知用户刷新token
	if newAuthUser != nil {
		userChanInsert(int64(0), userID, MsgTypeSys, ChildTypeRefreshToken, *newAuthUser)
	}

	// 消息处理
	switch clientMsg.MsgType {
	case MsgTypeUser:
		// 消息放入用户管道
		userChanInsert(
			clientMsg.Auth.ID,
			clientMsg.Target,
			clientMsg.MsgType,
			clientMsg.ChildType,
			clientMsg.Data)

	case MsgTypeSys:
		// 初始化用户管道
		if clientMsg.ChildType == ChildTypeSysInit {
			go func() {
				userChanList[userChanName] = make(chan string)
				log.Println("add user channel :", userChanList)

				for {
					chanMsg := <- userChanList[userChanName]
					// 退出监听
					if chanMsg == "close" {
						return
					}

					if chanMsg != "" {
						err := conn.WriteMessage(websocket.TextMessage, []byte(chanMsg))
						if err != nil {
							log.Println("ws write msg: ", err)
						}
					}
				}
			}()
		}
	}

	return userID
}

// 生成用户管道名称
func getChanName(userID int64) string {
	return fmt.Sprintf("chan_%v", userID)
}

// 向用户管道插入消息
func userChanInsert(fromID int64, targetID int64, msgType int, childType int, data interface{}) {
	if _,ok := userChanList[getChanName(targetID)]; ok{
		chanMsg := ChanMsg{
			MsgType: msgType,
			ChildType: childType,
			From: fromID,
			Data: data,
		}

		chanJson, err := json.Marshal(chanMsg)
		if err != nil {
			log.Println("channel msg to json:", err)
		} else {
			userChanList[getChanName(targetID)] <- string(chanJson)
		}
	} else {
		log.Println("user channel not exists", getChanName(targetID))
	}
}
