package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/wenchangshou2/zebus/pkg/e"
	"github.com/wenchangshou2/zebus/pkg/http_api"
	"github.com/wenchangshou2/zebus/pkg/logging"
	"github.com/wenchangshou2/zebus/pkg/safety"
	"github.com/wenchangshou2/zebus/pkg/setting"
	"github.com/wenchangshou2/zebus/pkg/utils"
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

//pingHandler 心跳
func (s *httpServer) pingHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	s.enableCors(&w, req)
	health := "ok"
	return health, nil
}

//getSystemMachineCode 获取系统硬件id
func (s *httpServer) getSystemMachineCode(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	var (
		uuid   string
		newStr string
		err    error
	)
	s.enableCors(&w, req)
	safety := safety.Safety{}
	safety.DefaultKey()
	if uuid, err = utils.GetSystemUUID(); err != nil {
		return nil, http_api.Err{Code: e.ERROR, Text: err.Error()}
	}
	msec := time.Now().UnixNano() / 1000000
	systemInfo := SystemMachineCode{
		Uuid:    uuid,
		Date:    msec,
		Service: "Zebus",
	}
	out, _ := json.Marshal(systemInfo)
	newStr, _ = safety.EncryptWithSha1Base64(string(out))
	return struct {
		Msg string `json:"code"`
	}{newStr}, nil
}

// 获取授权状态
func (s *httpServer) getAuthorizationStatus(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {

	s.enableCors(&w, req)
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

func (s *httpServer) doPUB(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	s.enableCors(&w, req)
	readMax := setting.AppSetting.MaxMsgSize + 1
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, readMax))
	if err != nil {
		return nil, http_api.Err{Code: 500, Text: "INTERNAL_ERROR"}
	}
	if int64(len(body)) == readMax {
		return nil, http_api.Err{Code: 413, Text: "MSG_TOO_BIG"}
	}
	if len(body) == 0 {
		return nil, http_api.Err{Code: 400, Text: "MSG_EMPTY"}
	}
	reqParams, client, topic, _, err := s.getTopicFromQuery(req)
	if err != nil {
		return nil, err
	}
	var deferred time.Duration
	if ds, ok := reqParams["defer"]; ok {
		var di int64
		di, err = strconv.ParseInt(ds[0], 10, 64)
		if err != nil {
			return nil, http_api.Err{Code: 400, Text: "INTERNAL_ERROR"}
		}
		deferred = time.Duration(di) * time.Millisecond
	}
	topic = strings.TrimPrefix(topic, "/zebus")
	msg := NewMessage(client.GenerateID(), body, []byte(topic))
	msg.deferred = deferred
	err = client.PutMessage(msg)
	if err != nil {
		return nil, http_api.Err{Code: 503, Text: "EXITING"}
	}
	return "OK", nil
}

// doPUBv3: 推送异步的调用
func (s *httpServer) doPUBv3(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	var (
		err error
		//postJob string
		job   e.Job
		bytes []byte
		body  []byte
		topic string
	)
	s.enableCors(&w, req)
	readMax := setting.AppSetting.MaxMsgSize + 1
	if body, err = ioutil.ReadAll(io.LimitReader(req.Body, readMax)); err != nil {
		return nil, http_api.Err{
			Code: 500,
			Text: "Internal_error",
		}
	}
	if topic, err = s.getTopic(req); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &job); err != nil {
		return nil, http_api.Err{Code: 500, Text: "parse json failed"}
	}
	job.Topic = topic

	fmt.Println("save job", job, bytes, GJobMgr)
	if len(job.Name) == 0 {
		job.Name = strconv.Itoa(int(time.Now().UnixNano()))
	}
	if _, err = GJobMgr.SaveJob(&job); err != nil {
		return nil, http_api.Err{Code: 500, Text: "save key error:" + err.Error()}
	}

	return "OK", nil
}

// mPub 多消息推送
func (s *httpServer) mPub(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	var (
		err     error
		readMax int64
		body    []byte
	)
	s.enableCors(&w, req)
	readMax = setting.AppSetting.MaxMsgSize + 1
	body, err = ioutil.ReadAll(io.LimitReader(req.Body, readMax))
	if err != nil {
		return nil, http_api.Err{Code: 500, Text: "INTERNAL_ERROR"}
	}
	if int64(len(body)) == readMax {
		return nil, http_api.Err{Code: 413, Text: "MSG_TOO_BIG"}
	}
	if len(body) == 0 {
		return nil, http_api.Err{Code: 400, Text: "MSG_EMPTY"}
	}
	return nil, nil
}
func (s *httpServer) doPUBv2(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	s.enableCors(&w, req)
	readMax := setting.AppSetting.MaxMsgSize + 1
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, readMax))
	if err != nil {
		return nil, http_api.Err{Code: 500, Text: "INTERNAL_ERROR"}
	}
	if int64(len(body)) == readMax {
		return nil, http_api.Err{Code: 413, Text: "MSG_TOO_BIG"}
	}
	if len(body) == 0 {
		return nil, http_api.Err{Code: 400, Text: "MSG_EMPTY"}
	}
	fmt.Println("body", string(body))
	reqParams, client, topic, timeOut, err := s.getTopicFromQuery(req)
	if err != nil {
		return nil, err
	}
	var deferred time.Duration
	if ds, ok := reqParams["defer"]; ok {
		var di int64
		di, err = strconv.ParseInt(ds[0], 10, 64)
		if err != nil {
			return nil, http_api.Err{Code: 400, Text: "INVLID_DEFER"}
		}
		deferred = time.Duration(di) * time.Millisecond
	}
	topic = strings.TrimPrefix(topic, "/zebus")
	msg := NewMessage(client.GenerateID(), body, []byte(topic))
	msg.deferred = deferred
	client.PutMessage((msg))
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Millisecond)
	done := client.AddNewWaitMessage(msg.ID)
	select {
	case messageBody, ok := <-done:
		if ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write(*messageBody)
		}
	case <-ctx.Done():
		w.Header().Set("Content-Type", "application/json")
		d, _ := json.Marshal(http_api.Err{Code: 500, Text: "No Message RECEIVED"})
		w.Write(d)
	}
	client.DeleteWitMessage(msg.ID)
	return nil, nil
}

// doPUBGroup :推送到分组
func (s *httpServer) doPUBGroup(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	s.enableCors(&w, req)
	readMax := setting.AppSetting.MaxMsgSize + 1
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, readMax))
	if err != nil {
		return nil, http_api.Err{
			Code: 500,
			Text: "INTERNAL_ERROR",
		}
	}
	if int64(len(body)) == readMax {
		return nil, http_api.Err{Code: 413, Text: "MSG_TOO_BIG"}
	}
	if len(body) == 0 {
		return nil, http_api.Err{Code: 400, Text: "MSG_EMPTY"}
	}
	return nil, nil
}

func newHTTPServer(zebusd *ZEBUSD, tlsEnabled bool, tlsRequired bool) (*httpServer, error) {
	log := http_api.Log(logging.G_Logger)
	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = http_api.LogPanicHandler(logging.G_Logger)
	router.NotFound = http_api.LogNotFoundHandler(logging.G_Logger)
	router.MethodNotAllowed = http_api.LogMethodNotAllowedHandler(logging.G_Logger)
	allowedHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		}
		//if r.Header.Get("Access-Control-Request-Method") != "" {
		//	header := w.Header()
		//	header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
		//	header.Set("Access-Control-Allow-Origin", "*")
		//}
		w.WriteHeader(http.StatusNoContent)
	})
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
	router.Handle("POST", "/pubV2", http_api.Decorate(s.doPUBv2, http_api.V1))
	router.Handle("POST", "/pubV3", http_api.Decorate(s.doPUBv3, http_api.V1))
	router.Handle("POST", "/mpub", http_api.Decorate(s.mPub, http_api.V1))
	router.Handle("POST", "/pubGroup", http_api.Decorate(s.doPUBGroup, http_api.V1))
	router.Handle("POST", "/getClient", http_api.Decorate(s.getClient, http_api.V1))
	router.Handle("GET", "/clients", http_api.Decorate(s.getClients, http_api.V1))
	router.GET("/ws", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		serveWs(zebusd, writer, request)
	})
	return s, nil
}
func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}
func (s *httpServer) enableCors(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
	(*w).Header().Set("Access-Control-Request-Headers", "*")
}

// getTopic: 获取参数里面的topic字段
func (s *httpServer) getTopic(req *http.Request) (string, error) {
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		s.ctx.logf.Error(fmt.Sprintf("failed to parse request params - %s", err.Error()))
		return "", err
	}
	topic, ok := reqParams["topic"]
	if !ok {
		return "", http_api.Err{Code: 400, Text: "MISSING_APP_TOPIC"}
	}
	return topic[0], nil
}

// getTopicFromQuery 解析url来获取相应topic
func (s *httpServer) getTopicFromQuery(req *http.Request) (url.Values, *Client, string, int, error) {
	var (
		timeOut int
	)
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		s.ctx.logf.Error(fmt.Sprintf("failed to parse request params - %s", err))
		return nil, nil, "", 0, http_api.Err{Code: 400, Text: "INVALID REQUEST"}
	}
	topicNames, ok := reqParams["topic"]
	if !ok {
		return nil, nil, "", 0, http_api.Err{Code: 400, Text: "MISSING_APG_TOPIC"}
	}
	topicName := topicNames[0]
	client := s.ctx.getClients(topicName)
	if client == nil {
		return nil, nil, "", 0, http_api.Err{Code: 400, Text: "DRIVER NOT ONLINE"}
	}
	timeOutStr, ok := reqParams["timeOut"]
	if !ok {
		timeOut = 2000
	} else {
		tmp, err := strconv.Atoi(timeOutStr[0])
		if err != nil {
			timeOut = 2000
		} else {
			timeOut = tmp
		}
	}
	return reqParams, client, topicName, timeOut, nil
}
func (s *httpServer) getIpFromQuery(req *http.Request) (url.Values, string, error) {
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		fmt.Println("err", err.Error())
		return nil, "", http_api.Err{
			Code: 400,
			Text: "INVALID REQUEST",
		}
	}
	ip, ok := reqParams["ip"]
	if !ok {
		return nil, "", http_api.Err{
			Code: 0,
			Text: "MISSING_APG_IP",
		}
	}
	return reqParams, ip[0], nil

}

//getClients 获取所有的客户端信息
func (s *httpServer) getClients(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	s.enableCors(&w, req)
	if setting.EtcdSetting.Enable {
		d := G_workerMgr.GetAllClientInfo(s.ctx.getOnlineServer())
		return d, nil
	} else {
		return s.ctx.GetAllClientInfo(), nil
	}
}

// getClient 获取客户信息
func (s *httpServer) getClient(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	_, ip, err := s.getIpFromQuery(req)
	if err != nil {
		return nil, err
	}
	if setting.EtcdSetting.Enable {
		d := G_workerMgr.GetClientFromIp(ip)
		return d, nil
	} else {

	}
	return nil, nil
}
