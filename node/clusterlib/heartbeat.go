package node

import (
	"fmt"
	"net"
	"net/rpc"
)

var serverclient *rpc.Client

func ConnectToServer(ip string) {
	LocalAddr, _ := net.ResolveTCPAddr("tcp", ":0")
	ServerAddr, _ := net.ResolveTCPAddr("tcp", ip)
	conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	if err != nil {
		fmt.Println("Could not connect to server")
	} else {
		fmt.Println("Connecting to server on:", conn.LocalAddr().String())
		serverclient = rpc.NewClient(conn)
	}
}

func ServerHeartBeat() {
	var _ignored bool
	serverclient.Call("TServer.HeartBeat", &_ignored, &_ignored)
}
