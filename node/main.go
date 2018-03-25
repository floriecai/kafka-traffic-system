package main

import (
	"time"
	"fmt"
	"net"
	"net/rpc"
	"os"

	"./clusterlib"
)

var NodeInstance *shared.Node

type ClusterRpc struct {
}

type PeerRpc struct {
}

/*******************************
| Cluster RPC Calls
********************************/
func ListenClusterRpc() {
	cRpc := new(ClusterRpc)
	server := rpc.NewServer()
	server.RegisterName("Cluster", cRpc)
	tcp, _ := net.Listen("tcp", ":0")
	fmt.Println("ClusterRpc is listening on: ", tcp.Addr().String())
	server.Accept(tcp)
}

/*******************************
| Peer RPC Calls
********************************/
func ListenPeerRpc() {
	pRpc := new(PeerRpc)
	server := rpc.NewServer()
	server.RegisterName("Peer", pRpc)
	tcp, _ := net.Listen("tcp", ":0")
	fmt.Println("PeerRpc is listening on: ", tcp.Addr().String())
	server.Accept(tcp)
}

/*******************************
| Helpers
********************************/

// Code from https://gist.github.com/jniltinho/9787946
func GeneratePublicIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error generating public IP")
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String() + ":"
			}
		}
	}

	return "Could not find IP"
}

func ManageConnections() {
	interval := time.Duration(NodeInstace.Settings.HeartBeat / 5)
	heartbeat := time.Tick(interval * time.Millisecond)
	count := 0

	for {
		select {
			case <- heartbeat:
				node.ServerHeartBeat()
			}
	}
}

/*******************************
| Main
********************************/
func main() {
	serverIP := os.Args[1]
	// Open Cluster to App RPC
	go ListenClusterRpc()
	// Open Peer to Peer RPC
	go ListenPeerRpc()

	// Set up node
	NodeInstance = new(shared.Node)
	NodeInstance.Type = 1
	publicIP = GeneratePublicIP()
	fmt.Println(publicIP)
	ln, _ := net.Listen("tcp", publicIP)
	addr := ln.Addr()
	NodeInstance.Addr = addr
	// dummy values
	coords := shared.GPSCoordinates{Lat: 0.0, Lon: 0.0}
	NodeInstance.Coordinates = coords
	NodeInstance.Peers = []shared.Node

	// Connect to the Server
	node.ConnectToServer(serverIP)
	NodeInstance.Settings = node.Register(addr)

	fmt.Printf("%+v\n", NodeInstance)

	// Channels
	go ManageConnections()
}
