package gin

import (
	"github.com/gin-gonic/gin"
	"os"
)

func ExampleGinLoggerHandler() {
	router := gin.New()
	router.Use(GinLoggerHandler(os.Stdout), gin.Recovery())
	// TODO: 绑定路由
	// 启动
	router.Run()
}
