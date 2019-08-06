package http_api

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"strings"
)

type logWriter struct {
	logf *zap.Logger
}

func (l logWriter) Write(p []byte) (int, error) {
	l.logf.Warn(fmt.Sprintf("%s", string(p)))
	return len(p), nil
}
func Serve(listener net.Listener, handler http.Handler, proto string, logf zap.Logger) error {
	logf.Info(fmt.Sprintf("%s: listening on %s", proto, listener.Addr()))
	server := &http.Server{
		Handler:  handler,
		ErrorLog: log.New(logWriter{&logf}, "", 0),
	}
	err := server.Serve(listener)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		return fmt.Errorf("http.Serve() error - %s", err)
	}
	logf.Info(fmt.Sprintf("%s: closing  %s", proto, listener.Addr()))
	return nil
}
