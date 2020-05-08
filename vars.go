package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	GoEnvTools "github.com/kaizer666/goenvtools"
	GoLogger "github.com/kaizer666/gologger"
	"github.com/kaizer666/serviceCommunicatorServer"
)

var (
	globalServices          servicesStruct
	telegram                telegramStruct
	servicesFileName        string
	servicesFile            *os.File
	environment             GoEnvTools.GoEnv
	logger                  GoLogger.Logger
	fileDescriptor          *int
	exit1                   = make(chan int)
	writeServiceStopChannel = make(chan int)
	checkServiceStopChannel = make(chan int)
	server                  serviceCommunicatorServer.ServerStruct
	httpClient              = http.Client{Transport: &transport, Timeout: 2 * time.Second}
	writeServiceChannel     = make(chan serviceStruct)
	transport               = http.Transport{
		DialContext:       dialContext,
		DisableKeepAlives: true,
	}
)

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: 1 * time.Second}
	return d.DialContext(ctx, network, addr)
}
