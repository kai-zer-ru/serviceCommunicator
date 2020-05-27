package main

import (
	"encoding/json"
	"flag"
	"fmt"
	GoEnvTools "github.com/kaizer666/goenvtools"
	"os"
	"runtime"
	"syscall"
)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fileDescriptor = flag.Int("fd", 0, "Server socket fileDescriptor")
	getServices := flag.Bool("get", false, "")
	flag.Parse()

	logFile, _ := os.OpenFile("panic.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, os.FileMode(0644))
	_ = syscall.Dup2(int(logFile.Fd()), 2)

	environment = GoEnvTools.GoEnv{}
	_ = environment.InitEnv()
	fmt.Printf("Process PID : %v\n", os.Getpid())
	err := initConfig()
	if err != nil {
		panic(err)
	}

	if *getServices {
		data, err := redisCache.Get(serviceCommunicatorData)
		if err != nil {
			fmt.Printf("Error while get services: %v", err)
			return
		}
		if data != nil {
			fmt.Println(data.(string))
			return
		}
		fmt.Println("No services data")
		return
	}

	globalServices = servicesStruct{}
	servicesData := map[string]serviceStruct{}
	data, err := redisCache.Get(serviceCommunicatorData)
	if err == nil {
		if data != nil {
			err = json.Unmarshal([]byte(data.(string)), &servicesData)
			if err != nil {
				logger.Error("error parse serviceCommunicatorData: %v", err)
				servicesData = map[string]serviceStruct{}
			}
		}
	}
	globalServices.Services = servicesData
	go signalListener()
	logger.Info("before writeService")
	go func() {
		writeService()
	}()
	logger.Info("before ping")
	ping()
}
