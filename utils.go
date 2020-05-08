package main

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"os/signal"
)

func initConfig() error {
	isDebug := environment.GetEnvBool("IS_DEBUG", false)
	var LogLevel int
	if isDebug {
		LogLevel = 1
	} else {
		LogLevel = environment.GetEnvInt("LOG_LEVEL", 7)
	}
	err := logger.SetLogLevel(LogLevel)
	if err != nil {
		return err
	}
	err = logger.Init()
	if err != nil {
		return err
	}
	servicesFileName = environment.GetEnvString("SERVICES_FILE_NAME", "services.json")
	logger.Debug("Config read")
	return nil
}

func convertServiceToMap(service *serviceStruct) (map[string]interface{}, error) {
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
	logger.Debug("signalListener start")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	for {
		select {
		case <-quit:
			server.GraceHandler()
			return
		}
	}
}

func writeService() {
	for {
		select {
		case service := <-writeServiceChannel:
			globalServices.Locker.Lock()
			delete(globalServices.Services, service.Name)
			globalServices.Services[service.Name] = service
			data, err := json.Marshal(globalServices.Services)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(servicesFileName, data, 0644)
			if err != nil {
				panic(err)
			}
			globalServices.Locker.Unlock()
		case <-writeServiceStopChannel:
			return
		}
	}
}
