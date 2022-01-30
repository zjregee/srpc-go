package registry

import (
	"sync"
	"time"
	"math/rand"
	"sync/atomic"
)

type Network struct {
	mu             sync.Mutex
	reliable       bool
	longDelays     bool                        // pause a long time on send on disabled connection
	longReordering bool                        // sometimes delay replies a long time
	servers        map[interface{}]*Server     // servers, by name
	connections    map[interface{}]interface{} // endname -> servername
	done           chan struct{} // closed when Network is cleaned up
	count          int32         // total RPC count, for statistics
	bytes          int64         // total bytes send, for statistics
}

func (rn *Network) Cleanup() {
	close(rn.done)
}

func (rn *Network) Reliable(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.reliable = yes
}

func (rn *Network) LongReordering(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.longReordering = yes
}

func (rn *Network) LongDelays(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.longDelays = yes
}

func (rn *Network) readEndnameInfo(endname interface{}) (enabled bool,
	servername interface{}, server *Server, reliable bool, longreordering bool,
) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	enabled = rn.enabled[endname]
	servername = rn.connections[endname]
	if servername != nil {
		server = rn.servers[servername]
	}
	reliable = rn.reliable
	longreordering = rn.longReordering
	return
}

func (rn *Network) isServerDead(endname interface{}, servername interface{}, server *Server) bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.enabled[endname] == false || rn.servers[servername] != server {
		return true
	}
	return false
}

func (rn *Network) processReq(req reqMsg) {
	enabled, servername, server, reliable, longreordering := rn.readEndnameInfo(req.endname)

	if enabled && servername != nil && server != nil {
		if reliable == false {
			// short delay
			ms := (rand.Int() % 27)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}

		if reliable == false && (rand.Int()%1000) < 100 {
			// drop the request, return as if timeout
			req.replyCh <- replyMsg{false, nil}
			return
		}

		// execute the request (call the RPC handler).
		// in a separate thread so that we can periodically check
		// if the server has been killed and the RPC should get a
		// failure reply.
		ech := make(chan replyMsg)
		go func() {
			r := server.dispatch(req)
			ech <- r
		}()

		// wait for handler to return,
		// but stop waiting if DeleteServer() has been called,
		// and return an error.
		var reply replyMsg
		replyOK := false
		serverDead := false
		for replyOK == false && serverDead == false {
			select {
			case reply = <-ech:
				replyOK = true
			case <-time.After(100 * time.Millisecond):
				serverDead = rn.isServerDead(req.endname, servername, server)
				if serverDead {
					go func() {
						<-ech // drain channel to let the goroutine created earlier terminate
					}()
				}
			}
		}

		// do not reply if DeleteServer() has been called, i.e.
		// the server has been killed. this is needed to avoid
		// situation in which a client gets a positive reply
		// to an Append, but the server persisted the update
		// into the old Persister. config.go is careful to call
		// DeleteServer() before superseding the Persister.
		serverDead = rn.isServerDead(req.endname, servername, server)

		if replyOK == false || serverDead == true {
			// server was killed while we were waiting; return error.
			req.replyCh <- replyMsg{false, nil}
		} else if reliable == false && (rand.Int()%1000) < 100 {
			// drop the reply, return as if timeout
			req.replyCh <- replyMsg{false, nil}
		} else if longreordering == true && rand.Intn(900) < 600 {
			// delay the response for a while
			ms := 200 + rand.Intn(1+rand.Intn(2000))
			// Russ points out that this timer arrangement will decrease
			// the number of goroutines, so that the race
			// detector is less likely to get upset.
			time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
				atomic.AddInt64(&rn.bytes, int64(len(reply.reply)))
				req.replyCh <- reply
			})
		} else {
			atomic.AddInt64(&rn.bytes, int64(len(reply.reply)))
			req.replyCh <- reply
		}
	} else {
		// simulate no reply and eventual timeout.
		ms := 0
		if rn.longDelays {
			// let Raft tests check that leader doesn't send
			// RPCs synchronously.
			ms = (rand.Int() % 7000)
		} else {
			// many kv tests require the client to try each
			// server in fairly rapid succession.
			ms = (rand.Int() % 100)
		}
		time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
			req.replyCh <- replyMsg{false, nil}
		})
	}

}

// func (rn *Network) AddServer(servername interface{}, rs *Server) {
// 	rn.mu.Lock()
// 	defer rn.mu.Unlock()

// 	rn.servers[servername] = rs
// }

// func (rn *Network) DeleteServer(servername interface{}) {
// 	rn.mu.Lock()
// 	defer rn.mu.Unlock()

// 	rn.servers[servername] = nil
// }

// connect a ClientEnd to a server.
// a ClientEnd can only be connected once in its lifetime.
func (rn *Network) Connect(endname interface{}, servername interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.connections[endname] = servername
}

// enable/disable a ClientEnd.
func (rn *Network) Enable(endname interface{}, enabled bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.enabled[endname] = enabled
}

// get a server's count of incoming RPCs.
func (rn *Network) GetCount(servername interface{}) int {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	svr := rn.servers[servername]
	return svr.GetCount()
}

func (rn *Network) GetTotalCount() int {
	x := atomic.LoadInt32(&rn.count)
	return int(x)
}

func (rn *Network) GetTotalBytes() int64 {
	x := atomic.LoadInt64(&rn.bytes)
	return x
}


type Server struct {
	server_name    string
	server_key     string
	server_ip      string
	server_port    string
	server_enabled bool
	lastActiveTime time.Time
	services       []*Service
}

type Service struct {
	service_name    string
	service_key     string
	method_name     string
	describtion     string
	service_enabled bool
}


func (rn *Network) Sync() {

}

func (rn *Network) checkTimeout() {
	for {		
		time.Sleep(200)
		rn.mu.Lock()
		for _, server := range rn.servers {
			server.server_enabled = false
		}
		rn.mu.Unlock()
	}
}

func (rn *Network) refreshTimeout(serverName string, isLocked bool) {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return
	}
	server.lastActiveTime = time.Now()
	server.server_enabled = true
}

func (rn *Network) checkServiceAtMostOne(serverName, serviceName, methodName string, isLocked bool) bool {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return true
	}
	services := []*Service{}
	for _, service := range server.services {
		if service.service_name == serviceName && service.method_name == methodName  {
			services = append(services, service)
		}
	}
	return len(services) <= 1
}

func (rn *Network) solveConflictServices() {
	
}

func (rn *Network) checkService(serverName, serviceName, methodName string, isLocked bool) bool {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	services := rn.getService(serverName, serviceName, methodName, true)
	if services != nil {
		return true
	}
	return false
}

func (rn *Network) checkServiceName(serverName, serviceName string, isLocked bool) bool {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	services := rn.getServicesByServiceName(serverName, serviceName, true)
	if services != nil {
		return true
	}
	return false
}

func (rn *Network) checkMethodName(serverName, methodName string, isLocked bool) bool {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	services := rn.getServicesByMethodName(serverName, methodName, true)
	if services != nil {
		return true
	}
	return false
}

func (rn *Network) getServicesByServiceName(serverName, serviceName string, isLocked bool) []*Service {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return nil
	}
	res := []*Service{}
	for _, service := range server.services {
		if service.service_name == serviceName {
			res = append(res, service)
		}
	}
	if len(res) != 0 {
		return res
	}
	return nil
}

func (rn *Network) getServicesByMethodName(serverName, methodName string, isLocked bool) []*Service {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return nil
	}
	res := []*Service{}
	for _, service := range server.services {
		if service.method_name == methodName {
			res = append(res, service)
		}
	}
	if len(res) != 0 {
		return res
	}
	return nil
}

func (rn *Network) getService(serverName, serviceName, methodName string, isLocked bool) *Service {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return nil
	}
	for _, service := range server.services {
		if service.service_name == serviceName && service.method_name == methodName  {
			return service
		}
	}
	return nil
}

func (rn *Network) getServiceKey(serverName, serviceName, methodName string, isLocked bool) string {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	service := rn.getService(serverName, serviceName, methodName, true)
	if service == nil {
		return ""
	}
	return service.service_key
}

func (rn *Network) getServiceIndex(serverName, key string, isLocked bool) int {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return -1
	}
	for i, service := range server.services {
		if service.service_key == key {
			return i
		}
	}
	return -1
}

func (rn *Network) deleteServiceAtIndex(serverName string, index int, isLocked bool) bool {
	if !isLocked {
		rn.mu.Lock()
		defer rn.mu.Unlock()
	}

	server, ok := rn.servers[serverName]
	if !ok {
		return false
	}
	if len(server.services) < index + 1 {
		return false
	}
	rn.servers[serverName].services = append(rn.servers[serverName].services[:index], rn.servers[serverName].services[index+1:]...)
	return true
}

func (rn *Network) AddServer(serverName string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	_, ok := rn.servers[serverName]
	if ok {
		return
	}
	rn.servers[serverName] = &Server{}
}

func (rn *Network) DeleteServer(serverName string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	_, ok := rn.servers[serverName]
	if !ok {
		return
	}
	delete(rn.servers, serverName)
}

func (rn *Network) AddService(serverName, serviceName, methodName string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	server, ok := rn.servers[serverName]
	if !ok {
		return
	}
	ok = rn.checkService(serverName, serviceName, methodName, true)
	if ok {
		return
	}
	server.services = append(server.services, &Service{})
}

func (rn *Network) DeleteService(serverName, serviceName, methodName string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	_, ok := rn.servers[serverName]
	if !ok {
		return
	}
	ok = rn.checkService(serverName, serviceName, methodName, true)
	if !ok {
		return
	}
	serviceKey := rn.getServiceKey(serverName, serviceName, methodName, true)
	index := rn.getServiceIndex(serverName, serviceKey, true)
	_ = rn.deleteServiceAtIndex(serverName, index, true)
}

func (rn *Network) ReceiveHeartBeat() {

}

func (rn *Network) PullServices() {

}

func (rn *Network) PullService() {

}

func (rn *Network) GetStatus() {

}