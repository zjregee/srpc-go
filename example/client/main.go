package main

import (
	"srpc"
)

func main() {
	end, _ := srpc.MakeEnd()
	end.Call()
}