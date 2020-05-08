package main

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
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
			servicesFile, err := os.Open(servicesFileName)
			if err != nil {
				if os.IsNotExist(err) {
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
			data, err := json.Marshal(globalServices.Services)
			if err != nil {
				panic(err)
			}
			_, err = servicesFile.Write(data)
			if err != nil {
				panic(err)
			}
			globalServices.Locker.Unlock()
		case <-writeServiceStopChannel:
			return
		}
	}
}
