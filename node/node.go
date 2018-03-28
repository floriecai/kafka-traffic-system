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
	if err := node.BecomeLeader(ips, PeerRpcAddr); err != nil {
		return err
	}

	for _, ip := range ips {
		req := node.FollowMeMsg{
			LeaderIp:    PublicIp,
			FollowerIps: ips,
		}

		followerClient, err := rpc.Dial("tcp", ip)
		if err != nil {
			return err
		}

		if err = followerClient.Call("PeerRpc.FollowMe", req, _ignored); err != nil {
			return err
		}

		FollowerMap[ip] = followerClient
	}

	return nil
}

// Leader -> Node rpc that sets the caller as this node's leader
func (c PeerRpc) FollowMe(msg node.FollowMeMsg, _ignored *string) error {
	err := node.FollowLeader(msg)
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

/*******************************
| Main
********************************/
func main() {
	serverIP := os.Args[1]

	// Initiate data structures
	FollowerMap = make(map[string]*rpc.Client)
	PublicIp = node.GeneratePublicIP()

	// Open Filesystem on Disk
	node.MountFiles()
	// Open Peer to Peer RPC
	go ListenPeerRpc()
	// Connect to the Server
	node.ConnectToServer(serverIP)
	// Open Cluster to App RPC
	ListenClusterRpc()
}
