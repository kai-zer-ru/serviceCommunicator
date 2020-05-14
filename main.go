package main

import (
	"encoding/json"
	"flag"
	"fmt"
	GoEnvTools "github.com/kaizer666/goenvtools"
	"io/ioutil"
	"os"
	"runtime"
	"syscall"
)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fileDescriptor = flag.Int("fd", 0, "Server socket fileDescriptor")
	flag.Parse()

	logFile, _ := os.OpenFile(fmt.Sprintf("panic.%v.log", os.Getpid()), os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, os.FileMode(0644))
	_ = syscall.Dup2(int(logFile.Fd()), 2)

	environment = GoEnvTools.GoEnv{}
	_ = environment.InitEnv()
	fmt.Printf("Process PID : %v\n", os.Getpid())
	err := initConfig()
	if err != nil {
		panic(err)
	}

	telegram = telegramStruct{}
	telegram.BotToken = environment.GetEnvString("TELEGRAM_TOKEN", "")
	servicesFileIsExist := true
	servicesFile, err = os.Open(servicesFileName)
	if err != nil {
		if os.IsNotExist(err) {
			servicesFileIsExist = false
			servicesFile, err = os.Create(servicesFileName)
			if err != nil {
				logger.Error("error Create servicesFile: %v", err)
				panic(err)
			}
		} else {
			logger.Error("error: %v", err)
			panic(err)
		}
	}
	b, err := ioutil.ReadAll(servicesFile)
	defer func() {
		_ = servicesFile.Close()
	}()
	if err != nil {
		logger.Error("error: %v", err)
		panic(err)
	}
	globalServices = servicesStruct{}
	servicesData := map[string]serviceStruct{}
	if servicesFileIsExist {
		err = json.Unmarshal(b, &servicesData)
		if err != nil {
			logger.Error("error: %v", err)
			servicesData = map[string]serviceStruct{}
		}
	}
	globalServices.Services = servicesData
	go signalListener()
	go writeService()
	logger.Info("before ping")
	ping()
}
