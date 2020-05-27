package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kaizer666/RedisLibrary"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

func initConfig() error {
	isDebug := environment.GetEnvBool("IS_DEBUG", false)
	var LogLevel int
	if isDebug {
		LogLevel = 1
	} else {
		LogLevel = environment.GetEnvInt("LOG_LEVEL", 7)
	}
	logFileName := environment.GetEnvString("LOG_FILE_NAME", "")
	logger.SetLogFileName(logFileName)
	err := logger.SetLogLevel(LogLevel)
	if err != nil {
		return err
	}
	err = logger.Init()
	if err != nil {
		return err
	}
	redisCache = RedisLibrary.RedisType{
		Host:     environment.GetEnvString("REDIS_HOST", "127.0.0.1"),
		Port:     environment.GetEnvUint32("REDIS_PORT", 6379),
		Password: environment.GetEnvString("REDIS_PASSWORD", ""),
		DB:       environment.GetEnvInt("REDIS_DATABASE", 0),
	}
	err = redisCache.Connect()
	if err != nil {
		return err
	}
	telegram = telegramStruct{}
	telegram.BotToken = environment.GetEnvString("TELEGRAM_TOKEN", "")
	logger.Debug("Config read")
	return nil
}

func convertServiceToMap(service *serviceStruct) (map[string]interface{}, error) {
	logger.Info("start %s", funcName())
	serviceMap := make(map[string]interface{})
	serviceString, err := json.Marshal(service)
	if err != nil {
		return serviceMap, err
	}
	err = json.Unmarshal(serviceString, &serviceMap)
	if err != nil {
		return serviceMap, err
	}
	delete(serviceMap, "address")
	addresses := serviceMap["Addresses"]
	commands := serviceMap["Commands"]
	delete(serviceMap, "Addresses")
	delete(serviceMap, "Commands")
	serviceMap["addresses"] = addresses
	serviceMap["commands"] = commands
	return serviceMap, nil
}

func signalListener() {
	logger.Info("start %s", funcName())
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Kill)
	for {
		select {
		case <-quit:
			data, err := json.Marshal(globalServices.Services)
			if err == nil {
				_, _ = redisCache.Set(serviceCommunicatorData, data)
			}
			server.GraceStop()
			return
		}
	}
}

func writeService() {
	logger.Info("start %s", funcName())
	for {
		select {
		case service := <-writeServiceChannel:
			globalServices.Locker.Lock()
			delete(globalServices.Services, service.Name)
			globalServices.Services[service.Name] = service
			data, err := json.Marshal(globalServices.Services)
			if err != nil {
				globalServices.Locker.Unlock()
				panic(err)
			}
			_, err = redisCache.Set(serviceCommunicatorData, data)
			if err != nil {
				globalServices.Locker.Unlock()
				panic(err)
			}
			globalServices.Locker.Unlock()
		case <-writeServiceStopChannel:
			return
		}
	}
}

func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func getServiceAddress(service *serviceStruct) string {
	var availableAddresses []string
	for address, available := range service.Addresses {
		logger.Debug("Check " + address)
		resp, err := http.Get(address + "/ping")
		if err != nil {
			logger.Error("error: %v", err)
			if available {
				go sendUnavailableService(service.Name, address)
				available = false
			}
		} else {
			logger.Debug("Close " + address)
			err = resp.Body.Close()
			if err != nil {
				logger.Error(fmt.Sprintf("Error close response %s/ping", address))
				logger.Error("error: %v", err)
			}
			if !available {
				go sendAvailableService(service.Name, address)
				available = true
			}
			availableAddresses = append(availableAddresses, address)
		}
	}

	rand.Seed(time.Now().Unix())
	var address string
	if len(availableAddresses) > 0 {
		address = availableAddresses[rand.Intn(len(availableAddresses))]
	}
	return address
}
