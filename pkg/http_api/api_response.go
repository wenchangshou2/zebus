package http_api

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type Decorator func(handler APIHandler) APIHandler
type APIHandler func(w http.ResponseWriter, r *http.Request, param httprouter.Params) (interface{}, error)
type Err struct {
	Code int
	Text string
}

func (e Err) Error() string {
	return e.Text
}
func Decorate(f APIHandler, ds ...Decorator) httprouter.Handle {
	decorated := f
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		decorated(w, req, ps)
	}
}
func Log(logf *zap.Logger) Decorator {
	return func(f APIHandler) APIHandler {
		return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) (i interface{}, e error) {
			start := time.Now()
			response, err := f(w, r, param)
			elapsed := time.Since(start)
			status := 200
			if e, ok := err.(Err); ok {
				status = e.Code
			}
			logf.Info(fmt.Sprintf("%d %s %s (%s) %s", status, r.Method, r.URL.RequestURI(), r.RemoteAddr, elapsed))
			return response, err
		}
	}
}
func RespondV1(w http.ResponseWriter, code int, data interface{}) {
	var response []byte
	var err error
	var isJSON bool

	if code == 200 {
		switch data.(type) {
		case string:
			response = []byte(data.(string))
		case []byte:
			response = data.([]byte)
		case nil:
			response = []byte{}
		default:
			isJSON = true
			response, err = json.Marshal(data)
			if err != nil {
				code = 500
				data = err
			}
		}
	}

	if code != 200 {
		isJSON = true
		response, _ = json.Marshal(struct {
			Message string `json:"message"`
		}{fmt.Sprintf("%s", data)})
	}
	if isJSON {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	//w.Header().Set("X-zebus-Content-Type", "nsq; version=1.0")
	w.WriteHeader(code)
	w.Write(response)
}
func PlainText(f APIHandler) APIHandler {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) (i interface{}, e error) {
		code := 200
		data, err := f(w, r, param)
		if err != nil {
			code = err.(Err).Code
			data = err.Error()
		}
		switch d := data.(type) {
		case string:
			w.WriteHeader(code)
			io.WriteString(w, d)
		case []byte:
			w.WriteHeader(code)
			w.Write(d)
		default:
			panic(fmt.Sprintf("unknown response type %T", data))
		}
		return nil, nil
	}
}
func V1(f APIHandler) APIHandler {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) (i interface{}, e error) {
		data, err := f(w, r, param)
		if err != nil {
			RespondV1(w, err.(Err).Code, err)
			return nil, nil
		}
		RespondV1(w, 200, data)
		return nil, nil
	}
}
func LogPanicHandler(logf *zap.Logger) func(w http.ResponseWriter, req *http.Request, p interface{}) {
	return func(w http.ResponseWriter, req *http.Request, p interface{}) {
		logf.Error(fmt.Sprintf("panic in HTTP handler - %s", p))
		Decorate(func(w http.ResponseWriter, r *http.Request, param httprouter.Params) (i interface{}, e error) {
			return nil, Err{500, "INTERNAL ERROR"}
		}, Log(logf), V1)(w, req, nil)
	}
}
func LogNotFoundHandler(logf *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		Decorate(func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
			return nil, Err{404, "NOT_FOUND"}
		}, Log(logf), V1)(w, req, nil)
	})
}
func LogMethodNotAllowedHandler(logf *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		Decorate(func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
			return nil, Err{405, "METHOD_NOT_ALLOWED"}
		}, Log(logf), V1)(w, req, nil)
	})
}
