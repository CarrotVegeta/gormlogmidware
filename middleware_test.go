package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"testing"
)

var User UserService

type UserService struct {
}

func (u *UserService) Hello(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "hello middleware",
		},
	)
}
func (u *UserService) HelloRegister(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "hello register middleware",
		},
	)
}
func (u *UserService) Write(ps map[string]any, c *gin.Context, responseBody *bytes.Buffer) {
	log.Printf("打印register日志：方法：%s,路由：%s,响应:%s", c.Request.Method, c.Request.URL.Path, responseBody)
}
func (u *UserService) WriterLog(ps map[string]any, c *gin.Context, responseBody *bytes.Buffer) {
	log.Printf("打印日志：方法：%s,路由：%s,响应:%s", c.Request.Method, c.Request.URL.Path, responseBody)
}

var r *gin.Engine

func TestMain(m *testing.M) {
	r = gin.Default()
	m.Run()
	err := r.Run(":8080")
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
}
func TestLoggerMiddleWare(t *testing.T) {
	//测试直接使用中间件方法
	r.GET("/hello", LoggerMiddleWare(User.WriterLog), User.Hello)
	//测试注册日志方法
	lm := NewLoggerMiddleware()
	r.Use(LoggerMiddleWare())
	lm.RegisterGinHandler(http.MethodGet, "/hello-register", &User)
	r.GET("/hello-register", User.HelloRegister)
}
