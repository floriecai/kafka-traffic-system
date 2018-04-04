package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"../structs"
	c "./concurrentlib"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// ERRORS
///////////////////////////////////////////////////////////////////////////////////////////////////

type AddressAlreadyRegisteredError string

func (e AddressAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("Server: node address already registered [%s]", string(e))
}

type UnknownNodeError string

func (e UnknownNodeError) Error() string {
	return fmt.Sprintf("Server: unknown node address heartbeat [%s]", string(e))
}

type DuplicateTopicNameError string

func (e DuplicateTopicNameError) Error() string {
	return fmt.Sprintf("Server: Topic: [%s] already exists", string(e))
}

type TopicDoesNotExistError string

func (e TopicDoesNotExistError) Error() string {
	return fmt.Sprintf("Server: Topic: [%s] does not exist", string(e))
}

type InsufficientNodesForCluster string

func (e InsufficientNodesForCluster) Error() string {
	return fmt.Sprintf("Server: There are not enough available nodes to form a topic's cluster")
}

// END OF ERRORS
///////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////
// DATA STRUCTURES
///////////////////////////////////////////////////////////////////////////////////////////////////

type TServer struct {
	TopicsFileLock sync.Mutex
}

type Config struct {
	NodeSettings   structs.NodeSettings `json:"node-settings"`
	RpcIpPort      string               `json:"rpc-ip-port"`
	MinClusterSize uint32               `json:"min-cluster-size"`
}

const (
	topicFile string = "./topics.json"
)

type AllNodes struct {
	sync.RWMutex
	all map[string]*structs.Node // unique ip:port idenitifies a node
}

var (
	tServer *TServer
	config  Config

	allNodes    = AllNodes{all: make(map[string]*structs.Node)}
	orphanNodes = c.Orphanage{Orphans: make([]structs.Node, 0)}
	topics      = c.TopicCMap{Map: make(map[string]structs.Topic)}

	errLog = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
)

func readConfigOrDie(path string) {
	file, err := os.Open(path)
	handleErrorFatal("config file", err)

	buffer, err := ioutil.ReadAll(file)
	handleErrorFatal("read config", err)

	err = json.Unmarshal(buffer, &config)
	handleErrorFatal("parse config", err)
}

// Register Nodes
func (s *TServer) Register(n string, nodeSettings *structs.NodeSettings) error {
	allNodes.Lock()
	defer allNodes.Unlock()

	for _, node := range allNodes.all {
		if node.Address == n {
			return AddressAlreadyRegisteredError(n)
		}
	}

	outLog.Println("Register::Connecting to address: ", n)
	localAddr, err := net.ResolveTCPAddr("tcp", ":0")
	checkError(err, "GetPeers:ResolvePeerAddr")

	nodeAddr, err := net.ResolveTCPAddr("tcp", n)
	checkError(err, "GetPeers:ResolveLocalAddr")

	conn, err := net.DialTCP("tcp", localAddr, nodeAddr)
	checkError(err, "GetPeers:DialTCP")

	client := rpc.NewClient(conn)

	// Add to orphan nodes
	orphanNodes.Append(structs.Node{
		Address: n,
		Client:  client})

	allNodes.all[n] = &structs.Node{
		Address:         n,
		Client:          client,
		RecentHeartbeat: time.Now().UnixNano()}

	go monitor(n, time.Millisecond*time.Duration(config.NodeSettings.HeartBeat))

	*nodeSettings = config.NodeSettings

	outLog.Printf("Got Register from %s\n", n)

	return nil
}

// Writes to disk any connections that have been made to the server along
// with their corresponding topics (if any)
func updateNodeMap(addr string, topicName string, server *TServer) error {
	_, err := os.OpenFile(topicFile, os.O_WRONLY, 0644)

	if err == nil {
		// f.Write
		return nil
	}

	return err
}

func monitor(k string, heartBeatInterval time.Duration) {
	time.Sleep(time.Second * 10)
	for {
		allNodes.Lock()
		if time.Now().UnixNano()-allNodes.all[k].RecentHeartbeat > int64(heartBeatInterval) {
			outLog.Printf("%s timed out\n", allNodes.all[k].Address)
			delete(allNodes.all, k)
			allNodes.Unlock()
			return
		}
		outLog.Printf("%s is alive\n", allNodes.all[k].Address)
		allNodes.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func (s *TServer) HeartBeat(addr string, _ignored *bool) error {
	allNodes.Lock()
	defer allNodes.Unlock()
	if _, ok := allNodes.all[addr]; !ok {
		return UnknownNodeError(addr)
	}

	allNodes.all[addr].RecentHeartbeat = time.Now().UnixNano()

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Producer API RPC
///////////////////////////////////////////////////////////////////////////////////////////////////

func (s *TServer) CreateTopic(topicName *string, topicReply *structs.Topic) error {
	// Check if there is already a Topic with the same name
	if _, ok := topics.Get(*topicName); !ok {
		if orphanNodes.Len >= config.MinClusterSize {
			orphanNodes.Lock()
			lNode := orphanNodes.Orphans[0]
			orphanNodes.Unlock()

			allNodes.Lock()
			node, exists := allNodes.all[lNode.Address]
			allNodes.Unlock()

			if !exists {
				log.Fatalf("Discrepancy in orphan nodes vs. Node Map. [%s] does not exist in NodeMap\n", lNode.Address)
			}

			orphanIps := make([]string, 0)

			orphanNodes.Lock()
			defer orphanNodes.Unlock()
			for i, orphan := range orphanNodes.Orphans {
				if i >= int(config.MinClusterSize) {
					break
				}

				orphanIps = append(orphanIps, orphan.Address)
			}

			var leaderClusterRpc string
			if err := node.Client.Call("Peer.Lead", orphanIps, &leaderClusterRpc); err != nil {
				errLog.Println("Node [%s] could not accept Leader position.", lNode.Address)
				return err
			}

			orphanNodes.DropN(int(config.MinClusterSize))

			topic := structs.Topic{
				TopicName:   *topicName,
				MinReplicas: config.NodeSettings.MinNumNodeConnections,
				Leaders:     []string{leaderClusterRpc},
				Followers:   orphanIps[1:]}

			topics.Set(*topicName, topic)
			*topicReply = topic
			return nil
		}

		return InsufficientNodesForCluster("")
	}

	return DuplicateTopicNameError(*topicName)
}

func (s *TServer) GetTopic(topicName *string, topicReply *structs.Topic) error {
	if topic, ok := topics.Get(*topicName); ok {
		*topicReply = topic
		return nil
	}

	return TopicDoesNotExistError(*topicName)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers for Leader promotion/demotion
///////////////////////////////////////////////////////////////////////////////////////////////////

func (s *TServer) AddTopicLeader(topicName *string, newLeader *structs.Node) (err error) {
	// topicI, ok := s.TopicMap.Load(topicName)
	// topicI.(structs.)
	// if ok {
	// 	topic.Leaders = append(topic.Leaders, newLeader)
	// 	return nil
	// } else {
	// 	return TopicDoesNotExistError(topicName)
	// }
	return nil
}

// Does not preserve order
func (s *TServer) RemoveTopicLeader(topicName *string, oldLeader *structs.Node) (err error) {
	// topic, ok := s.TopicMap.Load(topicName)
	// if ok {
	// 	// Find index of oldLeader in topic.Leaders
	// 	idx := -1
	// 	for i, v := range topic.Leaders {
	// 		if oldLeader == v {
	// 			idx = i
	// 		}
	// 	}

	// 	// Swap oldLeader with last Leader and
	// 	// Remove last element from slice
	// 	if idx < 0 {
	// 		return LeaderDoesNotExistError()
	// 	} else {
	// 		topic.Leaders[len(topic.Leaders)-1], topic.Leaders[idx] =
	// 			topic.Leaders[idx], topic.Leaders[len(topic.Leaders)-1]

	// 		topic.Leaders = topic.Leaders[:len(topic.Leaders)-1]

	// 		return nil
	// 	}
	// } else {
	// 	return TopicDoesNotExistError(topicName)
	// }
	return nil
}

func main() {
	//gob.Register(&net.TCPAddr{})

	// Pass in IP as command line argument
	// ip := os.Args[1] + ":0"
	path := flag.String("c", "", "Path to the JSON config")
	flag.Parse()

	if *path == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	readConfigOrDie(*path)

	rand.Seed(time.Now().UnixNano())

	// Set up Server RPC
	tServer = new(TServer)
	server := rpc.NewServer()
	server.Register(tServer)

	l, err := net.Listen("tcp", config.RpcIpPort)

	handleErrorFatal("listen error", err)
	outLog.Printf("Server started. Receiving on %s\n", config.RpcIpPort)

	if err != nil {
		fmt.Sprintln("Server: Error initializing RPC Listener")
		return
	}

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}
}

func handleErrorFatal(msg string, e error) {
	if e != nil {
		errLog.Fatalf("%s, err = %s\n", msg, e.Error())
	}
}

func checkError(err error, parent string) bool {
	if err != nil {
		errLog.Println(parent, ":: found error! ", err)
		return true
	}
	return false
}
