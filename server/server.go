package main

import (
	"fmt"
	"net/rpc"
	"os"
	"sync"

	"../shared"
)


///////////////////////////////////////////////////////////////////////////////////////////////////
// DATA STRUCTURES
///////////////////////////////////////////////////////////////////////////////////////////////////

type TServer struct {
	// Node -> shared.Topic
	NodeMap        sync.Map
	// TopicName -> shared.Topic
	TopicMap       sync.Map
	MinReplicasNum uint32
}

func main() {
	// Pass in IP as command line argument
	ip := os.Args[1] + ":0"

	// Set up Server RPC
	tserver := new(TServer)
	server := rpc.NewServer()
	server.Register(tserver)

	l, err := net.Listen("tcp", ip)

	if err != nil {
		fmt.Sprintf("Server: Error initializing RPC Listener")
		return
	}

	for {
		conn, _ := l.Accept()
		go server.ServeConn(conn)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Producer API RPC
///////////////////////////////////////////////////////////////////////////////////////////////////

func (s TServer) CreateTopic(topicName string) (topic shared.Topic, err error) {
	// Check if there is already a Topic with the same name
	if _, ok := s.TopicMap.Load(topicName); ok == nil {
		topic := new(shared.Topic)
		topic.TopicName = topicName
		topic.MinReplicasNum = s.MinReplicasNum
		topic.Leaders = []shared.Node{}
		topic.Followers = []shared.Node{}
		// TODO: Populate Leaders and Followers with actual Nodes

		s.TopicMap.Store(topic)

		return topic, nil
	} else {
		return {}, DuplicateTopicNameError(topicName)
	}
}

func (s TServer) GetTopic(topicName string) (topic shared.Topic, err error) {
	topic, ok := s.TopicMap.Load(topicName)
	if ok {
		return topic, nil
	} else {
		return {}, TopicDoesNotExistError(topicName)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers for Leader promotion/demotion
///////////////////////////////////////////////////////////////////////////////////////////////////

func (s TServer) AddTopicLeader(topicName string, newLeader shared.Node) (err error) {
	topic, ok := s.TopicMap.Load(topicName)
	if ok {
		topic.Leaders = append(topic.Leaders, newLeader)
		return nil
	} else {
		return TopicDoesNotExistError(topicName)
	}
}

// Does not preserve order
func (s TServer) RemoveTopicLeader(topicName, oldLeader shared.Node) (err error) {
	topic, ok := s.TopicMap.Load(topicName)
	if ok {
		// Find index of oldLeader in topic.Leaders
		idx := -1
		for i, v := range topic.Leaders {
			if oldLeader == v {
				idx = i
			}
		}

		// Swap oldLeader with last Leader and
		// Remove last element from slice
		if (idx < 0) {
			return LeaderDoesNotExistError()
		} else {
			topic.Leaders[len(topic.Leaders)-1], topic.Leaders[idx] =
			topic.Leaders[idx], topic.Leaders[len(topic.Leaders)-1]

			topic.Leaders = topic.Leaders[:len(topic.Leaders)-1]
		
			return nil
		}
	} else {
		return TopicDoesNotExistError(topicName)
	}
}
