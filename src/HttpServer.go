package main
import (
	"fmt"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wenchangshou2/zebus/src/routers"
)


func InitHttpServer(port int)(err error){

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	routers.InitRouter(r)

	go r.Run(fmt.Sprintf(":%d",port))
	return
}
