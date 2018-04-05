package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"../structs"
	"./clusterlib"
)

type ClusterRpc struct {
	WriteLock *sync.Mutex
	WriteId   uint
}

type PeerRpc struct {
}

const WRITE_TIMEOUT_SEC = 10 // Time to wait for RPC call to peer nodes to confirm write

var ClusterRpcAddr, PeerRpcAddr, PublicIp string

var id int = 0

/*******************************
| Cluster RPC Calls
********************************/
func ListenClusterRpc(ln net.Listener) {
	cRpc := ClusterRpc{
		WriteLock: &sync.Mutex{},
	}
	server := rpc.NewServer()
	server.RegisterName("Cluster", cRpc)
	ClusterRpcAddr = ln.Addr().String()
	fmt.Println("ClusterRpc is listening on: ", ClusterRpcAddr)

	server.Accept(ln)
}

func (c ClusterRpc) WriteToCluster(write structs.WriteMsg, resp *bool) error {

	c.WriteLock.Lock()
	writeId := c.WriteId
	c.WriteLock.Unlock()

	if node.NodeMode == node.Leader {
		node.PeerMap.MapLock.RLock()

		writesCh := make(chan bool, 30)

		numRequiredWrites := node.MinReplicas
		// Subtract 1 because Leader is counted in ClusterSize and only Followers confirm Writes
		maxFailures := node.ClusterSize - numRequiredWrites - 1
		writeVerdictCh := node.CountConfirmedWrites(writesCh, numRequiredWrites, maxFailures)

		go func() {
			for ip, peer := range node.PeerMap.Map {
				var writeConfirmed bool

				resp := node.PropagateWriteReq{
					Topic:      write.Topic,
					VersionNum: writeId,
					LeaderId:   PublicIp,
					Data:       write.Data,
				}

				writeCall := peer.PeerConn.Go("Peer.ConfirmWrite", resp, &writeConfirmed, nil)

				go func(wc *rpc.Call) {
					select {
					case w := <-wc.Done:
						if w.Error != nil {
							fmt.Println("Peer [%s] REJECTED write", ip)
							writesCh <- false
						}
					case <-time.After(WRITE_TIMEOUT_SEC):
						writesCh <- false
					}
				}(writeCall)
			}
		}()

		// Block on writeVerdictCh
		writeSucceed := <-writeVerdictCh
		node.PeerMap.MapLock.RUnlock()

		if writeSucceed {
			if err := node.WriteNode(write.Topic, write.Data, writeId); err != nil {
				return err
			}

			*resp = true
			return nil
		}

		return node.InsufficientConfirmedWritesError("")
	}
	log.Println("WriteToCluster:: Node is not a leader. Should not have received Write")
	return errors.New("Node is not a leader. Cannot send Write")
}

func (c ClusterRpc) ReadFromCluster(topic string, response *[]string) error {
	topicData, err := node.ReadNode(topic)
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
func (c PeerRpc) Lead(ips []string, clusterAddr *string) error {
	err := node.BecomeLeader(ips, PeerRpcAddr)
	*clusterAddr = ClusterRpcAddr
	return err
}

// Leader -> Node rpc that sets the caller as this node's leader
func (c PeerRpc) FollowMe(msg node.FollowMeMsg, _ignored *string) error {
	err := node.FollowLeader(msg, PeerRpcAddr)
	return err
}

// Leader -> Node rpc that tells followers of new joining nodes
func (c PeerRpc) AddFollower(msg node.ModFollowerListMsg, _ignored *string) error {
	err := node.ModifyFollowerList(msg, true)
	return err
}

// Leader -> Node rpc that tells followers of nodes leaving
func (c PeerRpc) RemoveFollower(msg node.ModFollowerListMsg, _ignored *string) error {
	err := node.ModifyFollowerList(msg, false)
	return err
}

// Follower -> Leader rpc that is used to join this leader's cluster
// Used during the election process when attempting to connect to this leader
func (c PeerRpc) Follow(ip string, _ignored2 *string) error {
	fmt.Println("Peer.Follow from:", ip)
	err := node.PeerAcceptThisNode(ip)
	return err
}

// Follower -> Follower rpc that is used by the caller to become a peer of this node
func (c PeerRpc) Connect(_ignored1 string, _ignored2 *string) error {
	//TODO:
	return nil
}

// Node -> Node RPC that is used to notify of liveliness
func (c PeerRpc) Heartbeat(ip string, reply *string) error {
	id++
	//fmt.Println("hb from:", ip, id)
	return node.PeerHeartbeat(ip, reply, id)
}

// Leader -> Follower RPC to commit write
func (c PeerRpc) ConfirmWrite(req node.PropagateWriteReq, writeOk *bool) error {
	if err := node.WriteNode(req.Topic, req.Data, req.VersionNum); err != nil {
		checkError(err, "ConfirmWrite")
		return err
	}

	*writeOk = true
	return nil
}

/*******************************
| Main
********************************/

// Args:
// serverIP
// dataPath - a valid, existing directory path that ends with /
func main() {
	serverIP := os.Args[1]
	dataPath := os.Args[2]

	PublicIp = node.GeneratePublicIP()
	fmt.Println("The public IP is:", PublicIp)
	// Listener for clients -> cluster
	ln1, _ := net.Listen("tcp", PublicIp+"0")

	// Listener for server and other nodes
	ln2, _ := net.Listen("tcp", PublicIp+"0")

	// Open Filesystem on Disk
	node.MountFiles(dataPath)
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

func checkError(err error, parent string) bool {
	if err != nil {
		log.Println(parent, ":: found error! ", err)
		return true
	}
	return false
}
