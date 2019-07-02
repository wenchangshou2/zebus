package main
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/cors"
	"github.com/wenchangshou2/zebus/src/routers"
)
func InitHttpServer(address string,port int)(err error){
	r:=gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	routers.InitRouter(r)
	go r.Run(fmt.Sprintf("%s:%d",address,port))
	return
}