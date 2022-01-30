package main

import (
	"srpc"
)


type Arith struct {

}

func (t *Arith) Mul(a, b int) int {
	return a * b
}

func (t *Arith) Div(a, b int) int {
	return a / b
}

func main() {
	a := Arith{}
	svc := srpc.MakeService(a)
	src, err := srpc.MakeServer()
	if err != nil {
		return
	}
	src.AddService(svc)
	src.Serve()
}