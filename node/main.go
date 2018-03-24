package main

import ("./clusterlib"
		"../structs"
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

func (c ClusterRpc) Write(write structs.WriteMsg, response *string) error {
	// TODO: Do we need WriteMsg.Id?

	err := node.WriteFile(write.Topic, write.Data)
	// TODO: Call Propagate
	return err	
}

func (c ClusterRpc) Read(topic string, response *[]string) error {
	// TODO: Do we need WriteMsg.Id?

	topicData, err := node.ReadFile(topic)
	*response = topicData
	return err
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
	// Open Filesystem
	node.MountFile()
	// Open Cluster to App RPC
	go ListenClusterRpc()
	// Open Peer to Peer RPC
	go ListenPeerRpc()
	// Connect to the Server
	node.ConnectToServer(serverIP);
}