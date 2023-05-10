package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

// 自定义一个结构体，实现 gin.ResponseWriter interface
type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

// 重写 Write([]byte) (int, error) 方法
func (w responseWriter) Write(b []byte) (int, error) {
	//向一个bytes.buffer中写一份数据来为获取body使用
	w.b.Write(b)
	//完成gin.Context.Writer.Write()原有功能
	return w.ResponseWriter.Write(b)
}

type Logger interface {
	Write(ps map[string]any, c *gin.Context, responseBody *bytes.Buffer)
}

var OperateLogger *LoggerMiddleware

type LoggerMiddleware struct {
	LoggerHandlerM map[string]map[string]Logger
}

func NewLoggerMiddleware() *LoggerMiddleware {
	OperateLogger = &LoggerMiddleware{
		LoggerHandlerM: make(map[string]map[string]Logger),
	}
	return OperateLogger
}
func (l *LoggerMiddleware) GetUrlByMethod(method string) map[string]Logger {
	if l.LoggerHandlerM == nil {
		l.LoggerHandlerM = map[string]map[string]Logger{}
	}
	if url, ok := l.LoggerHandlerM[method]; ok {
		return url
	}
	return nil
}
func (l *LoggerMiddleware) GetHandler(method, url string) Logger {
	urlM := l.GetUrlByMethod(method)
	if urlM == nil {
		return nil
	}
	return urlM[url]
}
func (l *LoggerMiddleware) SetMethodM(method string, m map[string]Logger) {
	l.LoggerHandlerM[method] = m
}
func (l *LoggerMiddleware) SetHandler(method, url string, logger Logger) {
	if urlM := l.GetUrlByMethod(method); urlM != nil {
		urlM[url] = logger
		l.SetMethodM(url, urlM)
		return
	}
	m := map[string]Logger{
		url: logger,
	}
	l.SetMethodM(method, m)
}
func (l *LoggerMiddleware) RegisterGinHandler(method, url string, logger Logger) {
	l.SetHandler(method, url, logger)
}
func LoggerMiddleWare(writers ...func(pa map[string]any, c *gin.Context, b *bytes.Buffer)) func(c *gin.Context) {
	return func(c *gin.Context) {
		//提取参数
		ps, err := parseParams(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, err)
			return
		}
		writer := responseWriter{
			c.Writer,
			bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer

		//// 执行下一个中间件或路由处理函数
		c.Next()
		//记录日志
		if len(writers) > 0 {
			for _, w := range writers {
				w(ps, c, writer.b)
			}
			return
		}
		if OperateLogger != nil {
			handler := OperateLogger.GetHandler(c.Request.Method, c.Request.URL.Path)
			if handler == nil {
				return
			}
			handler.Write(ps, c, writer.b)
		}
	}
}

func parseParams(c *gin.Context) (map[string]any, error) {
	ps := parseUrlRequest(c.Request.URL.RawQuery)
	if c.Request.Header["Content-Type"] != nil && strings.Index(c.Request.Header["Content-Type"][0], "application/json") >= 0 {
		s, _ := io.ReadAll(c.Request.Body)
		if len(s) > 0 {
			if err := json.Unmarshal(s, &ps); err != nil {
				return nil, fmt.Errorf("parse paramse body err:%v", err)
			}
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(s))
	}
	if err := c.Request.ParseForm(); err == nil {
		c.PostForm("")
		for k, v := range c.Request.PostForm {
			ps[k] = v[0]
		}
	}
	return ps, nil
}
func parseUrlRequest(query string) map[string]any {
	res := make(map[string]interface{})
	ps := strings.Split(query, "&")
	for _, p := range ps {
		vs := strings.Split(p, "=")
		if len(vs) != 2 {
			continue
		}
		res[vs[0]] = vs[1]
	}
	return res
}
