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

	"../shared"
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

///////////////////////////////////////////////////////////////////////////////////////////////////
// END OF ERRORS
///////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////
// DATA STRUCTURES
///////////////////////////////////////////////////////////////////////////////////////////////////

type TServer struct {
	// Node -> shared.Topic
	NodeMap sync.Map
	// TopicName -> shared.Topic
	TopicMap       sync.Map
	MinReplicasNum uint32
	TopicsFileLock sync.Mutex
}

type NodeSettings struct {
	MinNumNodeConnections uint8  `json:"min-num-node-connections"`
	HeartBeat             uint32 `json:"heartbeat"`
}

type Node struct {
	Address         net.Addr
	RecentHeartbeat int64
	IsLeader        bool
}

type NodeInfo struct {
	Address net.Addr
}

type Config struct {
	NodeSettings NodeSettings `json:"node-settings"`
	RpcIpPort    string       `json:"rpc-ip-port"`
}

const (
	topicFile string = "./topics.json"
)

type AllNodes struct {
	sync.RWMutex
	all map[string]*Node // unique ip:port identifies a node
}

var (
	tServer  *TServer
	config   Config
	allNodes AllNodes    = AllNodes{all: make(map[string]*Node)}
	errLog   *log.Logger = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog   *log.Logger = log.New(os.Stderr, "[serv] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
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
func (s *TServer) Register(n NodeInfo, nodeSettings *NodeSettings) error {
	allNodes.Lock()
	defer allNodes.Unlock()

	// if node, exists := allNodes.all[n.Address.String()]; exists {
	// 	return AddressAlreadyRegisteredError(node.Address.String())
	// }

	for _, node := range allNodes.all {
		if node.Address.Network() == n.Address.Network() &&
			node.Address.String() == n.Address.String() {
			return AddressAlreadyRegisteredError(n.Address.String())
		}
	}

	allNodes.all[n.Address.String()] = &Node{Address: n.Address}

	go monitor(n.Address.String(), time.Duration(config.NodeSettings.HeartBeat))

	*nodeSettings = config.NodeSettings

	outLog.Printf("Got Register from %s\n", n.Address.String())

	return nil
}

// Writes to disk any connections that have been made to the server along
// with their corresponding topics (if any)
func updateNodeMap(addr string, topicName string, server *TServer) error {
	// server.NodeMap.Load()
	_, err := os.OpenFile(topicFile, os.O_WRONLY, 0644)

	if err == nil {
		// f.Write
		return nil
	}

	return err
}

func monitor(k string, heartBeatInterval time.Duration) {
	for {
		allNodes.Lock()
		if time.Now().UnixNano()-allNodes.all[k].RecentHeartbeat > int64(heartBeatInterval) {
			outLog.Printf("%s timed out\n", allNodes.all[k].Address.String())
			delete(allNodes.all, k)
			allNodes.Unlock()
			return
		}
		outLog.Printf("%s is alive\n", allNodes.all[k].Address.String())
		allNodes.Unlock()
		time.Sleep(heartBeatInterval)
	}
}

func (s *TServer) HeartBeat(addr *string, _ignored *bool) error {
	allNodes.Lock()
	defer allNodes.Unlock()

	if _, ok := allNodes.all[*addr]; !ok {
		return UnknownNodeError(*addr)
	}

	allNodes.all[*addr].RecentHeartbeat = time.Now().UnixNano()

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Producer API RPC
///////////////////////////////////////////////////////////////////////////////////////////////////

func (s TServer) CreateTopic(topicName *string, topicReply *shared.Topic) error {
	// Check if there is already a Topic with the same name
	if _, ok := s.TopicMap.Load(*topicName); !ok {
		// topic := new(shared.Topic)

		topic := shared.Topic{
			TopicName:      *topicName,
			MinReplicasNum: s.MinReplicasNum,
			Leaders:        []shared.Node{},
			Followers:      []shared.Node{}}
		// TODO: Populate Leaders and Followers with actual Nodes

		// s.TopicMap.Store(topic)

		*topicReply = topic
		return nil
	}

	return DuplicateTopicNameError(*topicName)

}

func (s *TServer) GetTopic(topicName *string, topicReply *shared.Topic) error {
	if topic, ok := s.TopicMap.Load(*topicName); ok {
		*topicReply = topic.(shared.Topic)
		return nil
	}

	return TopicDoesNotExistError(*topicName)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers for Leader promotion/demotion
///////////////////////////////////////////////////////////////////////////////////////////////////

func (s *TServer) AddTopicLeader(topicName *string, newLeader *shared.Node) (err error) {
	// topicI, ok := s.TopicMap.Load(topicName)
	// topicI.(shared.)
	// if ok {
	// 	topic.Leaders = append(topic.Leaders, newLeader)
	// 	return nil
	// } else {
	// 	return TopicDoesNotExistError(topicName)
	// }
	return nil
}

// Does not preserve order
func (s *TServer) RemoveTopicLeader(topicName *string, oldLeader *shared.Node) (err error) {
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
		fmt.Sprintf("Server: Error initializing RPC Listener")
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
