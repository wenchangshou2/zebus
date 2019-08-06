package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/http_api"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/safety"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
	"net/http"
	"strings"
	"time"
)

type httpServer struct {
	hub         *Hub
	tlsEnabled  bool
	tlsRequired bool
	router      http.Handler
}
type SystemMachineCode struct {
	Date    int64
	Uuid    string
	Service string
}

func (s *httpServer) pingHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	health := "ok"
	return health, nil
}
func (s *httpServer) getSystemMachineCode(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	var (
		uuid   string
		newStr string
		err    error
	)
	safey := safety.Safety{}
	safey.DefaultKey()
	if uuid, err = utils.GetSystemUUID(); err != nil {
		return nil, http_api.Err{Code: e.ERROR, Text: "获取唯一id失败"}
	}
	msec := time.Now().UnixNano() / 1000000
	systemInfo := SystemMachineCode{
		Uuid:    uuid,
		Date:    msec,
		Service: "Zebus",
	}
	out, err := json.Marshal(systemInfo)
	newStr, err = safey.EncryptWithSha1Base64(string(out))
	return struct {
		Msg string `json:"msg"`
	}{newStr}, nil
}
func (s *httpServer) getAuthorizationStatus(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {

	if setting.RunningSetting.IsAuthorization {
		return struct {
			Status            bool   `json:"status"`
			AuthorizationCode string `json:"AuthorizationCode"`
		}{true, setting.RunningSetting.AuthorizationCode}, nil
	} else {
		return struct {
			Status bool `json:"status"`
		}{false}, nil
	}
}
func (s *httpServer) getClients(w http.ResponseWriter,req *http.Request,ps httprouter.Params)(interface{},error){
	if setting.EtcdSetting.Enable{
		tmpOnlineList,err:=G_workerMgr.ListWorkers()
		tmpOfflineList:=make([]string,0)
		allServer,err:=G_workerMgr.GetAllClient()
		if err==nil{
			for _,v:=range allServer{
				isOffline:=true
				for _,onlineClient:=range tmpOnlineList{
					if strings.Compare(v,onlineClient.Ip)==0{
						isOffline=false
					}
				}
				if isOffline{
					tmpOfflineList=append(tmpOfflineList,v)
				}
			}
		}
		return struct {
			Online []e.WorkerInfo `json:"online"`
			Offline []string `json:"offline"`
		}{tmpOnlineList,tmpOfflineList},nil
	}else {
		return s.hub.GetAllClientInfo(),nil
	}
}
func (s *httpServer)put(w http.ResponseWriter,req *http.Request,ps httprouter.Params)(interface{},error){

}
func newHTTPServer(hub *Hub, tlsEnabled bool, tlsRequired bool) *httpServer {
	log := http_api.Log(logging.G_Logger)
	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = http_api.LogPanicHandler(logging.G_Logger)
	router.NotFound = http_api.LogNotFoundHandler(logging.G_Logger)
	router.MethodNotAllowed = http_api.LogMethodNotAllowedHandler(logging.G_Logger)
	s := &httpServer{
		hub:         hub,
		tlsEnabled:  tlsEnabled,
		tlsRequired: tlsRequired,
		router:      router,
	}
	router.Handle("GET", "/ping", http_api.Decorate(s.pingHandler, log, http_api.PlainText))
	router.Handle("POST", "/getSystemMachineCode", http_api.Decorate(s.getSystemMachineCode, log, http_api.V1))
	router.Handle("POST","/getAuthorizationStatus",http_api.Decorate(s.getAuthorizationStatus,log,http_api.V1))
	router.Handle("POST","/getClients",http_api.Decorate(s.getClients,log,http_api.V1))
	return s
}
func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//if !s.tlsEnabled && s.tlsRequired {
	//	resp := fmt.Sprintf(`{"message": "TLS_REQUIRED", "https_port": %d}`,
	//		s.ctx.nsqd.RealHTTPSAddr().Port)
	//	w.Header().Set("X-NSQ-Content-Type", "nsq; version=1.0")
	//	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//	w.WriteHeader(403)
	//	io.WriteString(w, resp)
	//	return
	//}
	fmt.Println("serverhttp", w, req)

	s.router.ServeHTTP(w, req)
}
