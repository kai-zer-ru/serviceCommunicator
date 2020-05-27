package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kaizer666/serviceCommunicatorServer"
	"io"
	"io/ioutil"
	"net/http"
)

func sendCommand(w http.ResponseWriter, req *http.Request) {
	//sendCommandStruct
	logger.Info("start %s", funcName())
	if req.Method != "POST" {
		_, _ = io.WriteString(w, `{"error": "request method is not post"}`)
		return
	}
	decoder := json.NewDecoder(req.Body)
	var commandData = sendCommandStruct{}
	err := decoder.Decode(&commandData)
	if err != nil {
		logger.Error("error: %v", err)
		_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}
	if commandData.DaemonName == "" {
		_, _ = io.WriteString(w, `{"error": "no DaemonName in request"}`)
		return
	}
	if commandData.Command == "" {
		_, _ = io.WriteString(w, `{"error": "no Command in request"}`)
		return
	}
	daemon, ok := globalServices.Services[commandData.DaemonName]
	if !ok {
		_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "no daemon with name %s"}`, commandData.DaemonName))
		return
	}
	var address = ""
	var method = "GET"
	var requiredParams []string
	for _, command := range daemon.Commands {
		if command.Name == commandData.Command {
			address = getServiceAddress(&daemon)
			method = command.Method
			requiredParams = command.RequiredParams
			break
		}
	}
	if address == "" {
		_, _ = io.WriteString(w, `{"error": "no address in daemons"}`)
		return
	}
	for _, param := range requiredParams {
		if _, ok := commandData.Params[param]; !ok {
			_, _ = io.WriteString(w, `{"error": "no all required_params in request"}`)
			return
		}
	}
	if method == "GET" {
		urlAddress := address + "/" + commandData.Command + "?"
		for param, value := range commandData.Params {
			urlAddress += fmt.Sprintf("%s=%s&", param, value)
		}
		if commandData.NeedResponse {
			resp, err := http.Get(urlAddress)
			if err != nil {
				_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%v"}`, err))
				return
			}
			responseData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				_ = resp.Body.Close()
				_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%v"}`, err))
				return
			}
			_ = resp.Body.Close()
			_, _ = io.WriteString(w, string(responseData))
			return
		}
		go func() {
			resp, err := http.Get(urlAddress)
			if err != nil {
				logger.Error("error: %v", err)
				return
			}
			_ = resp.Body.Close()
		}()
		_, _ = io.WriteString(w, `{}`)
		return
	}
	urlAddress := address + "/" + commandData.Command
	if commandData.NeedResponse {
		req, err := http.NewRequest("POST", urlAddress, nil)
		if err != nil {
			logger.Error("error: %v", err)
			return
		}
		req.Header.Add("Content-Type", `application/json`)
		params, marshalErr := json.Marshal(commandData.Params)
		if marshalErr != nil {
			logger.Error("error: %v", marshalErr)
			return
		}
		req.Body = ioutil.NopCloser(bytes.NewBufferString(string(params)))
		resp, err := httpClient.Do(req)
		if err != nil {
			_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%v"}`, err))
			return
		}
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			_, _ = io.WriteString(w, fmt.Sprintf(`{"error": "%v"}`, err))
			return
		}
		_ = resp.Body.Close()
		_, _ = io.WriteString(w, string(responseData))
		return
	}
	go func() {
		req, err := http.NewRequest("POST", urlAddress, nil)
		if err != nil {
			logger.Error("error: %v", err)
			return
		}
		req.Header.Add("Content-Type", `application/json`)
		params, marshalErr := json.Marshal(commandData.Params)
		if marshalErr != nil {
			logger.Error("error: %v", marshalErr)
			return
		}
		req.Body = ioutil.NopCloser(bytes.NewBufferString(string(params)))
		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Error("error: %v", err)
			return
		}
		_ = resp.Body.Close()
	}()
	_, _ = io.WriteString(w, `{}`)
}
func deleteService(w http.ResponseWriter, req *http.Request) {
	logger.Info("start %s", funcName())
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

func deleteDaemon(w http.ResponseWriter, req *http.Request) {
	logger.Info("start %s", funcName())
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
	serviceAddress, ok := serviceData["address"]
	if !ok {
		_, _ = io.WriteString(w, `{"error": "no address in request"}`)
		return
	}
	_, ok = globalServices.Services[serviceName.(string)]
	if !ok {
		_, _ = io.WriteString(w, `{"error": "no service"}`)
		return
	}
	globalServices.Locker.Lock()
	for address := range globalServices.Services[serviceName.(string)].Addresses {
		if address == serviceAddress.(string) {
			delete(globalServices.Services[serviceName.(string)].Addresses, address)
			break
		}
	}
	globalServices.Locker.Unlock()
	_, _ = io.WriteString(w, `{"error": "null"}`)
}

func getService(w http.ResponseWriter, req *http.Request) {
	logger.Info("start %s", funcName())
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
	// проверяем адреса
	address := getServiceAddress(&service)

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
	logger.Info("start %s", funcName())
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
	logger.Info("start %s", funcName())
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
		logger.Debug("serviceData = %v", serviceData)
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
	logger.Info("start %s", funcName())
	server = serviceCommunicatorServer.ServerStruct{}
	server.SetLogger(logger.GetLogger())
	server.SetEnvironment(&environment)
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/registerService": registerService,
		"/getService":      getService,
		"/getServices":     getServices,
		"/deleteService":   deleteService,
		"/deleteDaemon":    deleteDaemon,
		"/sendCommand":     sendCommand,
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
			RequiredParams: []string{
				"name",
				"description",
				"address",
			},
		},
		{
			Name:        "getService",
			Description: "getService",
			Params: map[string]string{
				"name": "name of service",
			},
			Method: "GET",
			RequiredParams: []string{
				"name",
			},
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
			RequiredParams: []string{
				"name",
			},
		},
		{
			Name:        "deleteDaemon",
			Description: "deleteDaemon",
			Params: map[string]string{
				"name":    "name of daemon",
				"address": "address of daemon",
			},
			Method: "POST",
			RequiredParams: []string{
				"name",
				"address",
			},
		},
		{
			Name:        "sendCommand",
			Description: "sendCommand",
			Params: map[string]string{
				"daemon_name":   "name of daemon",
				"command":       "name of command",
				"params":        "params to send",
				"need_response": "need_response",
			},
			Method: "POST",
			RequiredParams: []string{
				"daemon_name",
				"command",
			},
		},
	}
	server.SetHandlers(handlers)
	server.FileDescriptor = fileDescriptor
	server.ExitListener = &exit1
	server.StopFunctions = []func(){}
	server.StopChannels = make([]*chan int, 0)
	server.StopChannels = append(server.StopChannels, &checkServiceStopChannel, &writeServiceStopChannel)
	var whereHost = environment.GetEnvString("DOCKER_ADDRESS", "0.0.0.0:8888")
	server.SetAddress(whereHost)
	server.StartServer()
}
