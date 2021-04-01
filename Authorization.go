package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/safety"
	"github.com/wenchangshou2/zebus/pkg/utils"
)

// AuthorizationInfo 认证信息
type AuthorizationInfo struct {
	UUID           string `json:"uuid"`          //设备标识
	Expire         int    `json:"expire"`        //过期时间
	ID             int    `json:"id"`            //当前检验的id
	IsVerify       bool   `json:"isVerify"`      //是否定时检验有效性
	Service        string `json:"service"`       //服务名称
	VerifyAddress  string `json:"verifyAddress"` //校验api地址
	VerifyCycle    int    `json:"verifyCycle"`   //检验周期
	LastVerifyTime int64  `json:"lastVerifyTime"`
	ErrorCount     int    `json:"errorCount"` //错误次数
}
type AuthorizationRequest struct {
	Uuid string
}
type AuthorizationProcess struct {
	Status bool
	s      safety.Safety
	init   bool
	uuid   string
}
type VerifyResponse struct {
	Action bool   `json:"action"`
	Srouce string `json:"source"`
}

func (a *AuthorizationProcess) readLincenseFile() (string, error) {
	file, err := os.Open("License.dat")
	if err != nil {
		return "", err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func (a *AuthorizationProcess) RequestBind(id int, uuid string) {

}
func (a *AuthorizationProcess) QueryAuthorization() bool {
	return a.Status
}

//RequestVerify :请求远程检验
func (a *AuthorizationProcess) RequestVerify(info *AuthorizationInfo) (bool, error) {
	var (
		data    []byte
		err     error
		req     *http.Request
		doneStr string
	)
	if data, err = json.Marshal(info); err != nil {
		return false, err
	}
	content, err := safety.G_Safety.EncryptWithSha1Base64(string(data))
	if err != nil {
		return false, err
	}
	if req, err = http.NewRequest("POST", "http://39.98.68.4:9091/api/v1/verify", bytes.NewBuffer([]byte(content))); err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/plan")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	response := e.HttpResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, errors.New("解析Response失败")
	}
	if doneStr, err = safety.G_Safety.DecryptWithSha1Base64(response.Data.(string)); err != nil {
		return false, err
	}
	verifyResponse := &VerifyResponse{}

	err = json.Unmarshal([]byte(doneStr), &verifyResponse)
	if err != nil {
		return false, nil
	}
	return verifyResponse.Action, nil
}

// writeLicense:写入授权
func (A *AuthorizationProcess) writeLicense(info *AuthorizationInfo) error {
	now := int64(time.Now().Unix())
	info.LastVerifyTime = now
	tmp, _ := json.Marshal(info)
	jmStr, err := safety.G_Safety.EncryptWithSha1Base64(string(tmp))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("License.dat", []byte(jmStr), 0644)
	return err

}
func (a *AuthorizationProcess) verify(info *AuthorizationInfo) {
	isSuccess, err := a.RequestVerify(info)
	if err != nil {
		a.Status = false
		info.ErrorCount++
		return
	}
	if isSuccess {
		info.ErrorCount = 0
		a.Status = true
	} else {
		info.ErrorCount++
		a.Status = false
	}
	a.writeLicense(info)
}
func (a *AuthorizationProcess) Loop() {
	var (
		c       string
		content string
		err     error
		info    = &AuthorizationInfo{}
	)
	for {
		now := int(time.Now().Unix())

		if !utils.IsExist("License.dat") {
			G_Authorization.Status = false
			logging.G_Logger.Warn("授权文件不存在")
			goto next
		}
		c, _ = a.readLincenseFile()
		content, err = a.s.DecryptWithSha1Base64(c)
		if err != nil {
			a.Status = false
			goto next
		}
		err = json.Unmarshal([]byte(content), &info)
		if err != nil {
			a.Status = false
			goto next
		}
		//TODO info.UUID为空的判断
		if info.UUID != a.uuid {
			logging.G_Logger.Warn("授权文件错误")
			a.Status = false
			continue
		}
		if info.LastVerifyTime == 0 {
			a.verify(info)
			goto next
		}
		if info.IsVerify {
			nextQueryTime := (int(info.LastVerifyTime) + info.VerifyCycle)
			if now > nextQueryTime {
				a.verify(info)
				goto next
			}
		}

	next:
		time.Sleep(5 * time.Second)
	}
}
func (a *AuthorizationProcess) Init() error {
	var (
		uuid string
		err  error
	)
	a.s = safety.Safety{}
	a.s.DefaultKey()
	a.init = false
	if uuid, err = utils.GetSystemUUID(); err != nil {
		return err
	}
	a.uuid = uuid
	return nil
}

var (
	G_Authorization = &AuthorizationProcess{}
)

func InitAuthorization(done chan bool) (err error) {
	if !utils.IsExist("License.dat") {
		G_Authorization.Status = false
		logging.G_Logger.Warn("授权文件不存在")
	}
	if err := G_Authorization.Init(); err != nil {
		return err
	}
	go G_Authorization.Loop()
	return nil
}
