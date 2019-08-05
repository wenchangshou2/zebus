package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ping(c *gin.Context){
	fmt.Println("pong")
	c.String(http.StatusOK, "Pong")
}