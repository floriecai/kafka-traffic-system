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

var ClusterRpcAddr, PeerRpcAddr string

/*******************************
| Cluster RPC Calls
********************************/
func ListenClusterRpc() {
	cRpc := new(ClusterRpc)
	server := rpc.NewServer()
	server.RegisterName("Cluster", cRpc)
	tcp, _ := net.Listen("tcp", ":0")

	ClustrRpcAddr = tcp.Addr().String()
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

	PeerRpcAddr = tcp.Addr().String()
	fmt.Println("PeerRpc is listening on: ", tcp.Addr().String())

	server.Accept(tcp)
}

// Server -> Node rpc that sets that node as a leader
// When it returns the node will have been established as leader
func (c PeerRpc) Lead(ips []string, _ignored string) error {
	err := node.BecomeLeader(ips)
	return err
}

// Leader -> Node rpc that sets the caller as this node's leader
func (c PeerRpc) FollowMe(){
	//TODO:
}

// Follower -> Leader rpc that is used to join this leader's cluster
func (c PeerRpc) Follow(){
	//TODO:
}

// Follower -> Follower rpc that is used by the caller to become a peer of this node
func (c PeerRpc) Connect(){
	//TODO:
}
/*******************************
| Main
********************************/
func main() {
	serverIP := os.Args[1]
	// Open Filesystem on Disk
	node.MountFiles()
	// Open Cluster to App RPC
	go ListenClusterRpc()
	// Open Peer to Peer RPC
	go ListenPeerRpc()
	// Connect to the Server
	node.ConnectToServer(serverIP);
}