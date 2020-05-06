package serviceCommunicator

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	GoEnvTools "github.com/kaizer666/goenvtools"
)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fileDescriptor = flag.Int("fd", 0, "Server socket fileDescriptor")
	flag.Parse()

	logFile, _ := os.OpenFile("panic.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, os.FileMode(0644))
	_ = syscall.Dup2(int(logFile.Fd()), 2)
	environment = GoEnvTools.GoEnv{}
	_ = environment.InitEnv()
	fmt.Printf("Process PID : %v\n", os.Getpid())
	initConfig()

	telegram = telegramStruct{}
	telegram.BotToken = environment.GetEnvString("TELEGRAM_TOKEN", "")
	servicesFileIsExist := true
	var err error
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
	go checkService()
	ping()
}

func checkService() {
	for {
		select {
		case <-checkServiceStopChannel:
			return
		case service := <-checkServiceChannel:
			addresses := service.Addresses
			for address, available := range addresses {
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
				}
				service.Addresses[address] = available
				time.Sleep(time.Second)
			}
			writeServiceChannel <- service
		}
	}
}

func checkServices() {
	logger.Debug("Start checkServices")
	globalServices.Locker.Lock()
	services := make(map[string]serviceStruct)
	for name, service := range globalServices.Services {
		services[name] = service
	}
	globalServices.Locker.Unlock()
	for _, service := range services {
		logger.Debug("service: %v", service)
		checkServiceChannel <- service
		logger.Debug("service: %v", service)
	}
	globalServices.Locker.Lock()
	data, err := json.Marshal(globalServices.Services)
	if err != nil {
		globalServices.Locker.Unlock()
		logger.Error("error: %v", err)
		return
	}
	err = ioutil.WriteFile(servicesFileName, data, 0777)
	globalServices.Locker.Unlock()
	if err != nil {
		logger.Error("Error write services file")
		logger.Error("error: %v", err)
	}
	logger.Debug("stop checkServices")

}
