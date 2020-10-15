//+build license

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/safety"
	"io/ioutil"
	"time"
)

func CheckLicense()error{
	f, _ := GetFullPath("License.db")
	if !IsExist(f){
		return errors.New("授权文件未存在")
	}
	data,err:=ioutil.ReadFile(f)
	if err!=nil{
		return err
	}
	str,err:=safety.G_Safety.DecryptWithSha1Base64(string(data))
	license:=e.License{}
	err=json.Unmarshal([]byte(str),&license)
	if err!=nil{
		return err
	}
	machineCode,err:=GetSystemUUID()
	if err!=nil{
		return err
	}
	if license.Machine!=machineCode{
		return errors.New("当前授权文件不能应用本台机器")
	}
	now:=time.Now().Unix()
	if now>license.EndDate{
		return errors.New("当前授权文件过期")
	}
	if now<license.StartDate{
		return errors.New("当前授权文件未生效")
	}
	fmt.Println(license.Machine,license.StartDate,license.EndDate)
	return nil
}
