package serviceCommunicator

import (
	"encoding/json"
	"os"
	"os/signal"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func initConfig() {
	logFileName := environment.GetEnvString("LOG_FILENAME", "main.log")
	isDebug := environment.GetEnvBool("IS_DEBUG", false)
	var LogLevel int
	if isDebug {
		LogLevel = 1
	} else {
		LogLevel = environment.GetEnvInt("LOG_LEVEL", 7)
	}
	logger.SetLogLevel(LogLevel)
	logger.SetLogFileName(logFileName)
	logger.Init()
	servicesFileName = environment.GetEnvString("SERVICES_FILE_NAME", "services.json")
	logger.Debug("Config readed")
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
		case <-time.After(10 * time.Second):
			checkServices()
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
			globalServices.Locker.Unlock()
		case <-writeServiceStopChannel:
			return
		}
	}
}
