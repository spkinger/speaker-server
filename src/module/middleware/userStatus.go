package middleware

import (
	"encoding/json"
	"github.com/spkinger/speaker-server/src/module/auth"
	"github.com/spkinger/speaker-server/src/module/config"
	"github.com/spkinger/speaker-server/src/module/pub"
	"net/http"
	"net/url"
)

// 用户登录态验证
func UserStatusCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", config.Config.HttpAllowOrigin) // 于本地测试时开启，以兼容localhost和服务端域名
		writer.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
		writer.Header().Set("content-type", "application/json")             //返回数据格式是json
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization") //允许请求header Authorization
		writer.Header().Set("Access-Control-Expose-Headers", "Authorization") //允许响应header Authorization

		// cors类型的不进入后面流程
		if request.Method == "OPTIONS" {
			return
		}

		u, err := url.Parse(request.RequestURI)
		if err != nil {
			pub.ErrRep(writer,"url不合法", map[string]interface{}{})
			return
		}

		if u.Path == "/register" || u.Path == "/login" || u.Path == "/wss" {
			// 不验证，转向下个Handler
			next.ServeHTTP(writer, request)
			return
		}

		checkRes, authUser := auth.CheckLogin(request)
		if !checkRes {
			pub.AuthErrRep(writer,"用户未登录")
			return
		}

		// 通知用户更新token
		if authUser != nil {
			authJson, err := json.Marshal(authUser)
			if err != nil {
				pub.ErrRep(writer,"用户状态异常", map[string]interface{}{})
				return
			}

			writer.Header().Set("Authorization", url.QueryEscape(string(authJson)))
		}

		// 通过验证，转向下个Handler
		next.ServeHTTP(writer, request)
	})
}