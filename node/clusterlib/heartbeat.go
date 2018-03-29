package node

import (
	"fmt"
	"net"
	"net/rpc"
	"time"
	"../../structs"
)

var ServerClient *rpc.Client
var MinConnections uint8
var HBInterval uint32

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
func ServerRegister(addr string) {	
	var resp structs.NodeSettings
	err := ServerClient.Call("TServer.Register", addr, &resp)
	if err != nil {
		fmt.Printf("Error in heartbeat::Register()\n%s", err)
	}
	MinConnections = resp.MinNumNodeConnections
	HBInterval = resp.HeartBeat
}

func ServerHeartBeat(addr string) {
	var _ignored bool
	fmt.Printf("starting server hb of: %d\n", HBInterval)
	interval := time.Duration(HBInterval / 2)
	heartbeat := time.Tick(interval * time.Millisecond)
	for {
		select {
		case <- heartbeat:
			ServerClient.Call("TServer.HeartBeat", addr, &_ignored)
		}
	}
}

