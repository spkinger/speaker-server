package auth

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/spkinger/speaker-server/src/module/config"
	"github.com/spkinger/speaker-server/src/module/model"
	"github.com/spkinger/speaker-server/src/module/pub"
	"log"
	"net/http"
	"net/url"
	"time"
)

type AuthUser struct {
	ID int64        `json:"id"`
	Name string     `json:"name"`
	NickName string `json:"nick_name"`
	Token string    `json:"token"`
	Time int64      `json:"time"`
}

// 注册
func Register(writer http.ResponseWriter, request *http.Request) {
	userName := request.PostFormValue("user_name")
	nickName := request.PostFormValue("nick_name")
	password := request.PostFormValue("passwd")
	log.Println("register in")


	if len(password) < 8 {
		pub.ErrRep(writer,"密码不能小于8位", map[string]interface{}{})
		return
	}

	if checkUserExists(userName, nickName) {
		pub.ErrRep(writer,"用户已存在请更换账户名或昵称", map[string]interface{}{})
		return
	}

	en := encryptPasswd(userName, password)
	user := createUser(userName, nickName, en)

	if user != nil {
		// 注册成功自动登录
		auth := getAuthUser(userName, password)

		if auth != nil {
			pub.SuccRep(writer,"注册成功", *auth)
		} else {
			pub.ErrRep(writer, "注册成功，登录失败", map[string]string{})
		}
	} else {
		pub.ErrRep(writer,"注册失败", map[string]interface{}{})
		return
	}
}

// 登录
func Login(writer http.ResponseWriter, request *http.Request) {
	userName := request.PostFormValue("user_name")
	password := request.PostFormValue("passwd")

	auth := getAuthUser(userName, password)

	if auth != nil {
		pub.SuccRep(writer,"登录成功", *auth)
	} else {
		pub.ErrRep(writer, "用户不存在或密码不正确", map[string]string{})
	}
}

// 创建用户
func createUser(name string, nickName string, password string) *model.User {
	user := model.User{
		Name: name,
		NickName: nickName,
		Passwd: password,
	}
	err := user.Create()
	if err == nil {
		return &user
	} else {
		log.Print("create user err:", err)
		return nil
	}
}

// 检测用户是否存在
func checkUserExists(name string, nickName string) bool {
	user := model.User{}
	user.GetByName(name)

	if user.ID != 0 {
		return true
	}

	user.GetByNickName(nickName)

	if user.ID != 0 {
		return true
	} else {
		return false
	}
}

// 密码加密
func encryptPasswd(userName string, passwd string) string {
	en := md5.Sum([]byte(fmt.Sprintf("passwd-spkeaker-%v-%v", userName, passwd)))
	return fmt.Sprintf("%x", en)
}

// 生成token
func generateToken(t int64, userName string, passwd string) string {
	token := md5.Sum([]byte(fmt.Sprintf("sSpeakerp-%v-%v-%v", t, userName, passwd)))
	return fmt.Sprintf("%x", token)
}

// 获取一个权限用户
func getAuthUser(userName string, password string) *AuthUser {
	user := model.User{}
	user.GetByName(userName)

	if user.ID == 0 {
		return nil
	}

	// 密码不正确
	en := encryptPasswd(user.Name, password)
	if en != user.Passwd {
		return nil
	}

	// 生成token
	t := time.Now().Unix()
	token := generateToken(t, user.Name, user.Passwd)

	return &AuthUser{
		ID: user.ID,
		Name: user.Name,
		NickName: user.NickName,
		Token: token,
		Time: t,
	}
}

// 获取权限user--加密后的密码
func getAuthUserEnPassword(userName string, enPassword string) *AuthUser {
	user := model.User{}
	user.GetByName(userName)

	if user.ID == 0 {
		return nil
	}

	// 密码不正确
	if enPassword != user.Passwd {
		return nil
	}

	// 生成token
	t := time.Now().Unix()
	token := generateToken(t, user.Name, user.Passwd)

	return &AuthUser{
		ID: user.ID,
		Name: user.Name,
		NickName: user.NickName,
		Token: token,
		Time: t,
	}
}

// 检查用户态是否有效
func CheckUser(user AuthUser) (bool, *AuthUser) {
	if user.Name == "" {
		return false, nil
	}

	// 查找用户
	userModel := model.User{}
	userModel.GetByName(user.Name)
	if userModel.ID == 0 {
		return false, nil
	}

	// 封禁用户判断
	if userModel.Status != 1 {
		return false, nil
	}

	// 验证token
	password := userModel.Passwd
	checkToken := generateToken(user.Time, user.Name, password)
	if checkToken != user.Token {
		return false, nil
	}

	// token过期处理
	var newUser *AuthUser
	t := time.Now().Unix()
	if t - user.Time > config.Config.TokenTimeOut {
		newUser = getAuthUserEnPassword(user.Name, password)
		if newUser == nil {
			return false, nil
		}
	}

	return true, newUser
}

// http中验证登陆态
func CheckLogin(request *http.Request) (bool, *AuthUser) {
	user := GetAuthUserRequest(request)
	if user == nil {
		return false, nil
	} else {
		return CheckUser(*user)
	}
}

// 从请求中获取AuthUser
func GetAuthUserRequest(request *http.Request) *AuthUser {
	user := AuthUser{}
	userEncode := request.Header.Get("Authorization")
	urlJson, err := url.QueryUnescape(userEncode)
	if err != nil {
		log.Print("url decode authUser failed:", err)
		return nil
	}

	err = json.Unmarshal([]byte(urlJson), &user)
	if err != nil {
		log.Print("json to authUser failed:", err)
		return nil
	} else {
		return &user
	}
}

// 请求中获取用户ID
func GetUidRequest(request *http.Request) int64 {
	user := GetAuthUserRequest(request)
	if user == nil {
		return 0
	} else {
		return user.ID
	}
}