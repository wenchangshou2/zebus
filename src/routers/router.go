package routers

import (
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/wenchangshou2/zebus/src/routers/api"
)

func InitRouter( r *gin.Engine){
	r.Use(static.Serve("/",static.LocalFile("./view",true)))
	r.GET("/ping",api.Ping)
}