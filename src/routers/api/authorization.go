package api

import (
	"github.com/gin-gonic/gin"
	"github.com/wenchangshou2/zebus/src/pkg/app"
	"github.com/wenchangshou2/zebus/src/pkg/e"
	"github.com/wenchangshou2/zebus/src/pkg/utils"
	"net/http"
	"time"
)
type SystemMachineCode struct{
	Date int64
	Uuid string
}
func GetSystemMachineCode(c *gin.Context){
	var (
		uuid string
		err error
	)
	appG:=app.Gin{C:c}
	if uuid,err=utils.GetSystemUUID();err!=nil{
		appG.Response(http.StatusInternalServerError,e.ERROR,nil)
		return
	}
	msec := time.Now().UnixNano() / 1000000
	systemInfo:=SystemMachineCode{
		Uuid:uuid,
		Date:msec,
	}
	//out,err:=json.Marshal(systemInfo)
	appG.Response(http.StatusOK,e.SUCCESS,systemInfo)
}
