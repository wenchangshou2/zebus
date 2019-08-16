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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type httpServer struct {
	ctx         *ZEBUSD
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
func (s *httpServer) getClients(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	if setting.EtcdSetting.Enable {
		tmpOnlineList, err := G_workerMgr.ListWorkers()
		tmpOfflineList := make([]string, 0)
		allServer, err := G_workerMgr.GetAllClient()
		if err == nil {
			for _, v := range allServer {
				isOffline := true
				for _, onlineClient := range tmpOnlineList {
					if strings.Compare(v, onlineClient.Ip) == 0 {
						isOffline = false
					}
				}
				if isOffline {
					tmpOfflineList = append(tmpOfflineList, v)
				}
			}
		}
		clientsConfigInfo := G_workerMgr.GetClientConfigInfo()
		resourcesInfo:=G_workerMgr.GetResourceInfo()
		fmt.Println("resourcesInfo",resourcesInfo)
		for k, onlineClient := range tmpOnlineList {
			if info, ok := clientsConfigInfo[onlineClient.Ip]; ok {
				tmpOnlineList[k].Config = *info
			}
			if resource,ok:=resourcesInfo[onlineClient.Ip];ok{
				tmpOnlineList[k].Resource=resource
			}else{
				tmpOnlineList[k].Resource=make([]string,0)
			}
		}
		return struct {
			Online  []e.WorkerInfo `json:"online"`
			Offline []string       `json:"offline"`
		}{tmpOnlineList, tmpOfflineList}, nil
	} else {
		return s.ctx.GetAllClientInfo(), nil
	}
}
func (s *httpServer) doPUB(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	readMax := setting.AppSetting.MaxMsgSize + 1
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, readMax))
	if err != nil {
		return nil, http_api.Err{500, "INTERNAL_ERROR"}
	}
	if int64(len(body)) == readMax {
		return nil, http_api.Err{413, "MSG_TOO_BIG"}
	}
	if len(body) == 0 {
		return nil, http_api.Err{400, "MSG_EMPTY"}
	}
	reqParams, client, topic, err := s.getTopicFromQuery(req)
	if err != nil {
		return nil, err
	}
	var deferred time.Duration
	if ds, ok := reqParams["defer"]; ok {
		var di int64
		di, err = strconv.ParseInt(ds[0], 10, 64)
		if err != nil {
			return nil, http_api.Err{400, "INVLID_DEFER"}
		}
		deferred = time.Duration(di) * time.Millisecond
	}
	msg := NewMessage(client.GenerateID(), body, []byte(topic))
	msg.deferred = deferred
	err = client.PutMessage(msg)
	if err != nil {
		return nil, http_api.Err{503, "EXITING"}
	}
	return "OK", nil
}
func newHTTPServer(zebusd *ZEBUSD, tlsEnabled bool, tlsRequired bool) *httpServer {
	log := http_api.Log(logging.G_Logger)
	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = http_api.LogPanicHandler(logging.G_Logger)
	router.NotFound = http_api.LogNotFoundHandler(logging.G_Logger)
	router.MethodNotAllowed = http_api.LogMethodNotAllowedHandler(logging.G_Logger)
	s := &httpServer{
		ctx:         zebusd,
		tlsEnabled:  tlsEnabled,
		tlsRequired: tlsRequired,
		router:      router,
	}
	router.Handle("GET", "/ping", http_api.Decorate(s.pingHandler, log, http_api.PlainText))
	router.Handle("POST", "/getSystemMachineCode", http_api.Decorate(s.getSystemMachineCode, log, http_api.V1))
	router.Handle("POST", "/getAuthorizationStatus", http_api.Decorate(s.getAuthorizationStatus, log, http_api.V1))
	router.Handle("POST", "/getClients", http_api.Decorate(s.getClients, log, http_api.V1))
	router.Handle("POST", "/pub", http_api.Decorate(s.doPUB, http_api.V1))
	return s
}
func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}
func (s *httpServer) getTopicFromQuery(req *http.Request) (url.Values, *Client, string, error) {
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		s.ctx.logf.Error(fmt.Sprintf("failed to parse request params - %s", err))
		return nil, nil, "", http_api.Err{400, "INVALID REQUEST"}
	}
	topicNames, ok := reqParams["topic"]
	if !ok {
		return nil, nil, "", http_api.Err{400, "MISSING_APG_TOPIC"}
	}
	topicName := topicNames[0]
	client := s.ctx.getClients(topicName)
	if client == nil {
		return nil, nil, "", http_api.Err{400, "DRIVER NOT ONLINE"}
	}
	return reqParams, client, topicName, nil
}
