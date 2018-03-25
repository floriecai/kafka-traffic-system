package node

import (
	"fmt"
	"net"
	"net/rpc"

	"../../shared"
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

// Not sure if this belongs here
func Register(nodeAddr net.Addr) shared.NodeSettings {
	reqArgs := shared.NodeInfo{Address: nodeAddr}
	var resp shared.NodeSettings
	err := serverclient.Call("TServer.Register", reqArgs, &resp)
	if err != nil {
		fmt.Println("Error registering Node to Server")
	}
	return resp
}

func ServerHeartBeat() {
	var _ignored bool
	serverclient.Call("TServer.HeartBeat", &_ignored, &_ignored)
}
