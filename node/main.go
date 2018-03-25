package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"

	"../structs"
	"./clusterlib"
)

type ClusterRpc struct {
}

type PeerRpc struct {
}

<<<<<<< HEAD
=======
var ClusterRpcAddr, PeerRpcAddr string

>>>>>>> 982173f3ba3035f1e232a66a0fd94bc868241475
/*******************************
| Cluster RPC Calls
********************************/
func ListenClusterRpc() {
	cRpc := new(ClusterRpc)
	server := rpc.NewServer()
	server.RegisterName("Cluster", cRpc)
	tcp, _ := net.Listen("tcp", ":0")

	ClusterRpcAddr = tcp.Addr().String()
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
func (c PeerRpc) Lead(ips []string, _ignored *string) error {
	err := node.BecomeLeader(ips, PeerRpcAddr)
	return err
}

// Leader -> Node rpc that sets the caller as this node's leader
func (c PeerRpc) FollowMe(msg node.FollowMeMsg, _ignored *string) error {
	err := node.FollowLeader(msg)
	return err
}

// Follower -> Leader rpc that is used to join this leader's cluster
func (c PeerRpc) Follow(_ignored1 string, _ignored2 *string) error {
	//TODO:
	return nil
}

// Follower -> Follower rpc that is used by the caller to become a peer of this node
func (c PeerRpc) Connect(_ignored1 string, _ignored2 *string) error {
	//TODO:
	return nil
}
/*******************************
| Main
********************************/
func main() {
	serverIP := os.Args[1]
	// Open Filesystem on Disk
	node.MountFiles()
	// Open Peer to Peer RPC
	go ListenPeerRpc()
	// Connect to the Server
<<<<<<< HEAD
	node.ConnectToServer(serverIP)
}
=======
	node.ConnectToServer(serverIP);
	// Open Cluster to App RPC
	ListenClusterRpc()
}
>>>>>>> 982173f3ba3035f1e232a66a0fd94bc868241475
