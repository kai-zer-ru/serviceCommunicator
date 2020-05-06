package serviceCommunicator

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/kaizer666/serviceCommunicatorServer"
)

func deleteService(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		_, _ = io.WriteString(w, `{"error": "request method is not post"}`)
		return
	}
	decoder := json.NewDecoder(req.Body)
	var serviceData = map[string]interface{}{}
	err := decoder.Decode(&serviceData)
	if err != nil {
		logger.Error("error: %v", err)
		_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}
	serviceName, ok := serviceData["name"]
	if !ok {
		_, _ = io.WriteString(w, `{"error": "no name in request"}`)
		return
	}
	_, ok = globalServices.Services[serviceName.(string)]
	if !ok {
		_, _ = io.WriteString(w, `{"error": "no service"}`)
		return
	}
	globalServices.Locker.Lock()
	delete(globalServices.Services, serviceName.(string))
	globalServices.Locker.Unlock()
	_, _ = io.WriteString(w, `{"error": "null"}`)
}

func getService(w http.ResponseWriter, req *http.Request) {
	serviceName := req.URL.Query()["name"][0]
	if serviceName == "" {
		_, _ = io.WriteString(w, `{"error": "no name in request"}`)
		return
	}
	service, ok := globalServices.Services[serviceName]
	if !ok {
		_, _ = io.WriteString(w, `{"error": "no service"}`)
		return
	}
	var availableAddresses []string
	for address, available := range service.Addresses {
		if available {
			availableAddresses = append(availableAddresses, address)
		}
	}
	rand.Seed(time.Now().Unix())
	var address string
	if len(availableAddresses) > 0 {
		address = availableAddresses[rand.Intn(len(availableAddresses))]
	}

	response := make(map[string]interface{})
	response["address"] = address
	response["name"] = service.Name
	response["commands"] = service.Commands
	data, err := json.Marshal(response)
	if err != nil {
		logger.Error("error: %v", err)
		_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}
	_, _ = io.WriteString(w, string(data))
}

func getServices(w http.ResponseWriter, _ *http.Request) {
	services := globalServices.Services
	var response []map[string]interface{}
	for _, service := range services {
		serviceMap, err := convertServiceToMap(&service)
		if err != nil {
			logger.Error("error: %v", err)
			continue
		}
		response = append(response, serviceMap)
	}
	data, err := json.Marshal(response)
	if err != nil {
		logger.Error("error: %v", err)
		_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}
	_, _ = io.WriteString(w, string(data))
}

func registerService(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		_, _ = io.WriteString(w, `{"error": "request method is not post"}`)
		return
	}
	decoder := json.NewDecoder(req.Body)
	var serviceData = serviceStruct{}
	err := decoder.Decode(&serviceData)
	if err != nil {
		logger.Error("73 string: %v", err)
		_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}
	logger.Debug("serviceData = %v", serviceData)
	globalServices.Locker.Lock()
	currentService, ok := globalServices.Services[serviceData.Name]
	globalServices.Locker.Unlock()
	logger.Debug("ok = %v", ok)
	logger.Debug("currentService = %v", currentService)
	if !ok {
		currentAddress := serviceData.Address
		resp, err := http.Get(currentAddress + "/getCommands")
		if err != nil {
			logger.Error("83 string: %v", err)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error("90 string: %v", err)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			return
		}
		logger.Debug(string(body))
		var commands []serviceCommunicatorServer.CommandStruct
		err = json.Unmarshal(body, &commands)
		if err != nil {
			logger.Error("error: %v", err)
			_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			return
		}
		serviceData.Commands = commands
		serviceData.Addresses = make(map[string]bool)
		serviceData.Addresses[serviceData.Address] = true
		logger.Debug("serviceData = %v",serviceData)
		writeServiceChannel <- serviceData
		_, _ = io.WriteString(w, `{"error": null}`)
		return
	}
	logger.Debug("currentService.Addresses before = ")
	logger.Debug("currentService.Addresses = %v", currentService.Addresses)
	currentService.Addresses[serviceData.Address] = true
	logger.Debug("currentService.Addresses after = ")
	logger.Debug("currentService.Addresses = %v", currentService.Addresses)
	writeServiceChannel <- currentService
	_, _ = io.WriteString(w, `{"error": null}`)
	return
}

func ping() {
	server = serviceCommunicatorServer.ServerStruct{}
	server.SetLogger(logger.GetLogger())
	server.SetEnvironment(&environment)
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/registerService": registerService,
		"/getService":      getService,
		"/getServices":     getServices,
		"/deleteService":   deleteService,
	}
	server.Commands = []serviceCommunicatorServer.CommandStruct{
		{
			Name:        "getCommands",
			Description: "Список команд",
			Params:      map[string]string{},
			Method:      "GET",
		},
		{
			Name:        "registerService",
			Description: "registerService",
			Params: map[string]string{
				"name":        "name of service",
				"description": "description of service",
				"address":     "address of service",
			},
			Method: "POST",
		},
		{
			Name:        "getService",
			Description: "getService",
			Params: map[string]string{
				"name": "name of service",
			},
			Method: "GET",
		},
		{
			Name:        "getServices",
			Description: "getServices",
			Params:      map[string]string{},
			Method:      "GET",
		},
		{
			Name:        "deleteService",
			Description: "deleteService",
			Params: map[string]string{
				"name": "name of service",
			},
			Method: "POST",
		},
	}
	server.SetHandlers(handlers)
	server.FileDescriptor = fileDescriptor
	server.ExitListener = &exit1
	server.StopFunctions = []func(){}
	server.StopChannels = make([]*chan int, 0)
	server.StopChannels = append(server.StopChannels, &checkServiceStopChannel, &writeServiceStopChannel)
	var whereHost = environment.GetEnvString("WHERE_HOST", "")
	port := environment.GetEnvString("PING_PORT", "8888")
	server.SetAddress(fmt.Sprintf("%s:%s", whereHost, port))
	server.StartServer()
}
