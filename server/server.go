package server

import (
	"bufio"
	"log"
	"net"
	"srpc/common/protocol"
	"srpc/common/service"
	"strings"
	"sync"
)

type Registry struct {
	Registry_ip      string
	Registry_port    string
	Registry_enabled bool
}

type Server struct {
	mu       sync.Mutex
	services map[string]*service.Service
	count    int // incoming RPCs
	registry *Registry
	config   *Config
}

func MakeServer() (*Server, error){
	rs := &Server{}
	rs.services = map[string]*service.Service{}
	return rs, nil
}

func MakeServerFromConfig(fName string) (*Server, error) {
	var err error
	rs := &Server{}
	rs.services = map[string]*service.Service{}
	rs.config, err = NewConfig(fName, &JSONConfigFormat{})
	if err != nil {
		return nil, err
	}
	rs.registry = rs.config.Format.TransferToRegistry()
	return rs, nil
}

func MakeServerFromConfigText(text string) (*Server, error) {
	var err error
	rs := &Server{}
	rs.services = map[string]*service.Service{}
	rs.config, err = NewConfigFromText(text, &JSONConfigFormat{})
	if err != nil {
		return nil, err
	}
	rs.registry = rs.config.Format.TransferToRegistry()
	return rs, nil
}

func (rs *Server) RefreshConfig(fName string) error {
	var err error
	rs.config, err = NewConfig(fName, &JSONConfigFormat{})
	if err != nil {
		return err
	}
	rs.registry = rs.config.Format.TransferToRegistry()
	return nil
}

func (rs *Server) RefreshConfigFromText(text string) error {
	var err error
	rs.config, err = NewConfigFromText(text, &JSONConfigFormat{})
	if err != nil {
		return err
	}
	rs.registry = rs.config.Format.TransferToRegistry()
	return nil
}

func (rs *Server) AddService(svc *service.Service) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.services[svc.Name] = svc
}

func (rs *Server) Serve() {
	rs.InitWithConfigFile()
	rs.processReq([]byte{})

	listen, err := net.Listen("tcp", "127.0.0.1:20000")
	if err != nil {
		log.Fatalf("listen failed, err: %v", err)
		return
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalf("accept failed, err: %v", err)
			continue
		}
		go rs.process(conn)
	}

}

func (rs *Server) GetCount() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.count
}

func (rs *Server) process(conn net.Conn) {
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		var buf [128]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			log.Fatalf("read from client failed, err: %v", err)
			break
		}
		recvStr := string(buf[:n])
		log.Fatalf("收到client端发来的数据：%v", recvStr)
		conn.Write(rs.processReq(buf[:n]))
	}
}

func (rs *Server) processReq(data []byte) []byte {
	// 将字节流转换为请求
	rs.dispatch(protocol.ReqMsg{})
	return []byte{}
}

func (rs *Server) dispatch(req protocol.ReqMsg) protocol.ReplyMsg {
	rs.mu.Lock()

	rs.count += 1

	// split Raft.AppendEntries into service and method
	dot := strings.LastIndex(req.SvcMeth, ".")
	serviceName := req.SvcMeth[:dot]
	methodName := req.SvcMeth[dot+1:]

	service, ok := rs.services[serviceName]

	rs.mu.Unlock()

	if ok {
		return service.Dispatch(methodName, req)
	} else {
		choices := []string{}
		for k := range rs.services {
			choices = append(choices, k)
		}
		log.Fatalf("labrpc.Server.dispatch(): unknown service %v in %v.%v; expecting one of %v\n",
			serviceName, serviceName, methodName, choices)
		return protocol.ReplyMsg{Ok: false, Reply: nil}
	}
}

func (rs *Server) InitWithConfigFile() {
	rs.register()
	go rs.heartbeat()
}

// 

func (rs *Server) register() {
	
}

func (rs *Server) heartbeat() {

}
