package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/wenchangshou2/zebus/pkg/app"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/safety"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
	"net/http"
	"time"
)
type SystemMachineCode struct{
	Date int64
	Uuid string
	Service string
}
func GetSystemMachineCode(c *gin.Context){
	var (
		uuid string
		newStr string
		err error
		rtu map[string]interface{}
	)
	rtu=make(map[string]interface{})
	safey:=safety.Safety{}
	safey.DefaultKey()
	appG:=app.Gin{C:c}
	if uuid,err=utils.GetSystemUUID();err!=nil{
		appG.Response(http.StatusInternalServerError,e.ERROR,nil)
		return
	}
	msec := time.Now().UnixNano() / 1000000
	systemInfo:=SystemMachineCode{
		Uuid:uuid,
		Date:msec,
		Service:"Zebus",
	}
	out,err:=json.Marshal(systemInfo)
	newStr,err=safey.EncryptWithSha1Base64(string(out))
	rtu["msg"]=newStr
	appG.Response(http.StatusOK,e.SUCCESS,rtu)
}

func GetAuthorizationStatus(c *gin.Context){
	var (
		rtu map[string]interface{}
	)
	appG:=app.Gin{C:c}

	rtu=make(map[string]interface{})
	rtu["status"]=setting.RunningSetting.IsAuthorization
	if setting.RunningSetting.IsAuthorization {
		rtu["AuthorizationCode"] = setting.RunningSetting.AuthorizationCode
	}
	appG.Response(http.StatusOK,e.SUCCESS,rtu)
}