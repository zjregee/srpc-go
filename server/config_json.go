package server

import (
	"os"
	"strings"
	"encoding/json"
)

type JSONConfigFormat struct {
	Configuation_name    string `json:"configuation_name"`
	Configuration_key    string `json:"configuation_key"`
	Configuation_version string `json:"configuation_version"`
	Registry_ip          string `json:"registry_ip"`
	Registry_port        string `json:"registry_port"`
}

func (c *JSONConfigFormat) TransferToRegistry() *Registry {
	r := &Registry {
		Registry_ip: c.Registry_ip,
		Registry_port: c.Registry_port,
		Registry_enabled: true,
	}
	return r
}

func (c *JSONConfigFormat) TransferToFormat(rn *Registry) {
	
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