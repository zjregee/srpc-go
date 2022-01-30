package registry

import (
	"fmt"
	"net/http"
)

var rn *Network

func MakeNetwork() {
	rn = &Network{}
	rn.reliable = true
	rn.ends = map[interface{}]*ClientEnd{}
	rn.enabled = map[interface{}]bool{}
	rn.servers = map[interface{}]*Server{}
	rn.connections = map[interface{}](interface{}){}
	rn.endCh = make(chan reqMsg)
	rn.done = make(chan struct{})

	// single goroutine to handle all ClientEnd.Call()s
	go func() {
		for {
			select {
			case xreq := <-rn.endCh:
				atomic.AddInt32(&rn.count, 1)
				atomic.AddInt64(&rn.bytes, int64(len(xreq.args)))
				go rn.processReq(xreq)
			case <-rn.done:
				return
			}
		}
	}()
}

func Run() {
	if rn == nil {
		MakeNetwork()
	}
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}

func Close() {

}


func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
  }
  