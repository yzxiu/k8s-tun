package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yzxiu/k8s-tun/pkg/info"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&MyFormatter{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	server := info.Server()
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Errorf("server start error")
		}
	}()
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Errorf("server shutdown error")
	}
}

type MyFormatter struct{}

func (m *MyFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	var newLog string

	//HasCaller()为true才会有调用信息
	if entry.HasCaller() {
		fName := filepath.Base(entry.Caller.File)
		newLog = fmt.Sprintf("[%s] [%s] [%s:%d %s] %s\n",
			timestamp, entry.Level, fName, entry.Caller.Line, entry.Caller.Function, entry.Message)
	} else {
		newLog = fmt.Sprintf("[%s] [%s] %s\n", timestamp, entry.Level, entry.Message)
	}

	b.WriteString(newLog)
	return b.Bytes(), nil
}
