package main

import (
	"encoding/json"
	"sync"

	"github.com/kaizer666/serviceCommunicatorServer"
)

type sendCommandStruct struct {
	DaemonName   string                 `json:"daemon_name"`
	Command      string                 `json:"command"`
	Params       map[string]interface{} `json:"params,omitempty"`
	NeedResponse bool                   `json:"need_response"`
}

func (c *sendCommandStruct) toString() string {
	data, err := json.Marshal(c)
	if err != nil {
		logger.Error(err.Error())
		return ""
	}
	return string(data)
}

type serviceStruct struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Addresses   map[string]bool
	Address     string `json:"address,omitempty"`
	Commands    []serviceCommunicatorServer.CommandStruct
}

type servicesStruct struct {
	Services map[string]serviceStruct
	Locker   sync.Mutex
}
