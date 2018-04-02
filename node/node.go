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

var ClusterRpcAddr, PeerRpcAddr, PublicIp string

var FollowerMap map[string]*rpc.Client // ipAddr -> rpcClient

/*******************************
| Cluster RPC Calls
********************************/
func ListenClusterRpc(ln net.Listener) {
	cRpc := new(ClusterRpc)
	server := rpc.NewServer()
	server.RegisterName("Cluster", cRpc)
	ClusterRpcAddr = ln.Addr().String()
	fmt.Println("ClusterRpc is listening on: ", ClusterRpcAddr)

	server.Accept(ln)
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
func ListenPeerRpc(ln net.Listener) {
	pRpc := new(PeerRpc)
	server := rpc.NewServer()
	server.RegisterName("Peer", pRpc)
	PeerRpcAddr = ln.Addr().String()
	fmt.Println("PeerRpc is listening on: ", PeerRpcAddr)

	go server.Accept(ln)
}

// Server -> Node rpc that sets that node as a leader
// When it returns the node will have been established as leader
func (c PeerRpc) Lead(ips []string, _ignored *string) error {
	err := node.BecomeLeader(ips, PeerRpcAddr)
	return err
}

// Leader -> Node rpc that sets the caller as this node's leader
func (c PeerRpc) FollowMe(LeaderIp string, _ignored *string) error {
	err := node.FollowLeader(LeaderIp)
	return err
}

// Leader -> Node rpc that tells followers of new joining nodes
func (c PeerRpc) AddFollower(msg []string, _ignored *string) error {
	return nil
}

// Leader -> Node rpc that tells followers of nodes leaving
func (c PeerRpc) RemoveFollower(msg []string, _ignored *string) error {
	return nil
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

// Node -> Node RPC that is used to notify of liveliness
func (c PeerRpc) Heartbeat(ip string, reply *string) error {
	return node.PeerHeartbeat(ip, reply)
}

/*******************************
| Main
********************************/
func main() {
	serverIP := os.Args[1]

	// Initiate data structures
	FollowerMap = make(map[string]*rpc.Client)

	PublicIp = node.GeneratePublicIP()
	fmt.Println("The public IP is:", PublicIp)
	// Listener for clients -> cluster
	ln1, _ := net.Listen("tcp", PublicIp+"0")

	// Listener for server and other nodes
	ln2, _ := net.Listen("tcp", PublicIp+"0")

	// Open Filesystem on Disk
	node.MountFiles()
	// Open Peer to Peer RPC
	ListenPeerRpc(ln2)
	// Connect to the Server
	node.ConnectToServer(serverIP)
	node.ServerRegister(PeerRpcAddr)
	// Start Server Heartbeat
	go node.ServerHeartBeat(PeerRpcAddr)
	// Open Cluster to App RPC
	ListenClusterRpc(ln1)
}
