package pub

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type RepBody struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

// 构建响应
func CommonRep(code int, writer http.ResponseWriter, msg string, data interface{}) {
	rep := RepBody{
		Code: code,
		Msg: msg,
		Data: data,
	}

	repJson, err := json.Marshal(rep)
	if err != nil {
		log.Println("repBody to json: ", err)
		return
	}

	_, err = writer.Write(repJson)
	if err != nil {
		log.Println("write response: ", err)
	}
}

// 权限验证失败响应
func AuthErrRep(writer http.ResponseWriter, msg string)  {
	CommonRep(4000, writer, msg, map[string]interface{}{})
}

// 构造错误响应
func ErrRep(writer http.ResponseWriter, msg string, data interface{}) {
	CommonRep(10000, writer, msg, data)
}

// 成功响应
func SuccRep(writer http.ResponseWriter, msg string, data interface{}) {
	CommonRep(0, writer, msg, data)
}

// 切片中查找字符串
func StrInSlice(s string, list []string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}
	return false
}

// 获取get参数
func GetParam(request *http.Request, key string) (string, bool) {
	query := request.URL.Query()

	param, ok := query[key]
	if !ok {
		return "", ok
	}

	return strings.Join(param, ""), true
}

// 获取page参数组
func GetPageParam(writer http.ResponseWriter, request *http.Request) (int, int) {
	var err error
	pageDefault := 1
	pageSizeDefault := 25
	pageInt := pageDefault
	pageSizeInt := pageSizeDefault
	page, ok := GetParam(request, "page")
	if ok {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			log.Print("page to int err:", err)
			pageInt = pageDefault
		}
	}
	pageSize, ok := GetParam(request, "page_size")
	if ok {
		pageSizeInt, err = strconv.Atoi(pageSize)
		if err != nil {
			log.Print("page_size to int err:", err)
			pageSizeInt = pageSizeDefault
		}
	}
	if pageSizeInt < 10 {
		pageSizeInt = 10
	}
	if pageSizeInt > 200 {
		pageSizeInt = 200
	}

	return pageInt, pageSizeInt
}