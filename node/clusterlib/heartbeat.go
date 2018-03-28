package node

import (
	"fmt"
	"net"
	"net/rpc"

	"../../shared"
)

var ServerClient *rpc.Client
var MinConnections int
var HBInterval int

func ConnectToServer(ip string) {
	LocalAddr, _ := net.ResolveTCPAddr("tcp", ":0")
	ServerAddr, _ := net.ResolveTCPAddr("tcp", ip)
	conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	if err != nil {
		fmt.Println("Could not connect to server")
	} else {
		fmt.Println("Connecting to server on:", conn.LocalAddr().String())
		ServerClient = rpc.NewClient(conn)
	}
}

// Not sure if this belongs here
func Register(nodeAddr net.Addr) {	
	var resp structs.NodeSettings
	err := serverclient.Call("TServer.Register", nodeAddr, &resp)
	if err != nil {
		fmt.Printf("Error in heartbeat::Register()\n%s", err)
	}
	MinConnections = int(resp.MinNumNodeConnections)
	HeartBeatInterval = int(resp.HeartBeat)
}

func ServerHeartBeat() {
	var _ignored bool
	ServerClient.Call("TServer.HeartBeat", &_ignored, &_ignored)
}


func HeartBeatManager() {
	interval := time.Duration(NodeInstance.Settings.HeartBeat / 2)
	heartbeat := time.Tick(interval * time.Millisecond)
	// count := 0

	for {
		select {
		case <-heartbeat:
			node.ServerHeartBeat()
		}
	}
}

