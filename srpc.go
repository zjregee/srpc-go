package srpc

import (
	"srpc/server"
	"srpc/client"
	"srpc/registry"
	"srpc/common/service"
)

type ServerInterface interface {

}

type ClientInterface interface {

}

type RegistryInterface interface {
	
}

// Server Inteface

type Server struct {
	svr *server.Server
}

func (s *Server) RegisterName() {

}

func (s *Server) AddService(svc *Service) {

}

func (s *Server) Serve() {
	s.svr.Serve()
}

func (s *Server) Close() {

}

func MakeServer() (*Server, error) {
	server, err := server.MakeServer()
	if err != nil {
		return nil, err
	}
	return &Server{server}, nil
}

func MakeServerFromConfig(fName string) (*Server, error) {
	server, err := server.MakeServerFromConfig(fName)
	if err != nil {
		return nil, err
	}
	return &Server{server}, nil
}

func MakeServerFromConfigText(text string) (*Server, error) {
	server, err := server.MakeServerFromConfigText(text)
	if err != nil {
		return nil, err
	}
	return &Server{server}, nil
}

type Service struct {
	svc *service.Service
}

func MakeService(rcvr interface{}) *Service {
	return &Service{service.MakeService(rcvr)}
}

// Client Interface

type Client struct {
	end *client.ClientEnd
}

func (c *Client) RefrshConfig(fname string) error {
	return c.end.RefrshConfig(fname)
}

func (c *Client) RefreshConfigFromText(text string) error {
	return c.end.RefreshConfigFromText(text)
}

func (c *Client) Call() {

}

func (c *Client) Close() {
	c.end.Close()
}

func MakeEnd() (*Client, error) {
	client, err := client.MakeClientEnd()
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func MakeEndFromConfig(fName string) (*Client ,error) {
	client, err := client.MakeClientEndFromConfig(fName)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func MakeEndFromConfigText(text string) (*Client ,error) {
	client, err := client.MakeClientEndFromConfigText(text)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

// Registry Interface

type Registry struct {
	rn *registry.Network
}

func (r *Registry) Run() {

}

func MakeRegistry() *Registry {
	return &Registry{}
}