package main

import ("./clusterlib"
		"fmt"
		"net/rpc"
		"net"
		"os"
)

type ClusterRpc struct{

}

type PeerRpc struct{

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

// TODO: Decide on message struct
func (c ClusterRpc) Read(coords string, response *string) error {
	
}

func (c ClusterRpc) Write(req int, response *string) error {

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
| Main
********************************/
func main() {
	serverIP := os.Args[1]
	// Open Cluster to App RPC
	go ListenClusterRpc()
	// Open Peer to Peer RPC
	go ListenPeerRpc()
	// Connect to the Server
	node.ConnectToServer(serverIP);
}