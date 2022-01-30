package client

import (
	"os"
	"strings"
	"encoding/json"
)

type JSONConfigFormat struct {
	Configuation_name    string `json:"configuation_name"`
	Configuration_key    string `json:"configuation_key"`
	Configuation_version string `json:"configuation_version"`
	Server 				 []struct {
		Server_name string `json:"server_name"`
		Server_key  string `json:"server_key"`
		Server_ip   string `json:"server_ip"`
		Server_port string `json:"server_port"`
		Services 	[]struct {
			Service_name string `json:"service_name"`
			Service_key  string `json:"service_key"`
			Method_name  string `json:"method_name"`
			Describtion  string `json:"describtion"`
			Enabled      bool   `json:"enabled"`
		} `json:"services"`
	} `json:"server"`
}

func (c *JSONConfigFormat) TransferToNetWork() *Network {
	services := []*Service{}
	for _, server := range c.Server {
		for _, service := range server.Services {
			s := &Service {
				Service_name: service.Service_name,
				Method_name: service.Method_name,
				Server_ip: server.Server_ip,
				Server_port: server.Server_port,
				Service_enabled: true,
			}
			services = append(services, s)
		}		
	}
	rn := &Network {
		Services: services,
	}
	return rn
}

func (c *JSONConfigFormat) TransferToFormat(rn *Network) {
	
}

func (c *JSONConfigFormat) Write(fname string) error {
	filePtr, err := os.Create(fname)
    if err != nil {
        return err
    }
    defer filePtr.Close()
	return json.NewEncoder(filePtr).Encode(c)
}

func (c *JSONConfigFormat) Parse(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&c)
}

func (c *JSONConfigFormat) ParseFromText(text string) error {
	return json.NewDecoder(strings.NewReader(text)).Decode(&c)
}