package concurrentlib

import (
	"log"
	"net"
	"net/rpc"
	"sync"

	"../../structs"
)

// This file is reserved for server's maps and arry helpers to ensure thread-safety

type Orphanage struct {
	sync.RWMutex
	Orphans []structs.Node
	Len     uint32
}

// Map Structures
// To make it concurrent safe, we make a struct of each map, with a corresponding RWLock
// Naming convention: XCMap, where X is the map type, C stands for Concurrent
type NodeInfo struct {
	Ip   net.Addr
	Conn *rpc.Client
}

type NodeCMap struct {
	MapLock sync.RWMutex
	Map     map[string]NodeInfo // IpAddress -> Conn
}

type TopicCMap struct {
	MapLock sync.RWMutex
	Map     map[string]structs.Topic // Topicname -> Topic
}

func (nm *NodeCMap) Get(k string) (NodeInfo, bool) {
	nm.MapLock.RLock()
	defer nm.MapLock.RUnlock()
	v, exists := nm.Map[k]
	return v, exists
}

func (nm *NodeCMap) Set(k string, v NodeInfo) {
	nm.MapLock.Lock()
	defer nm.MapLock.Unlock()
	nm.Map[k] = v
}

func (tm *TopicCMap) Get(k string) (structs.Topic, bool) {
	tm.MapLock.RLock()
	defer tm.MapLock.RUnlock()
	v, exists := tm.Map[k]
	return v, exists
}

func (tm *TopicCMap) Set(k string, v structs.Topic) {
	tm.MapLock.Lock()
	defer tm.MapLock.Unlock()
	tm.Map[k] = v
}

func (o *Orphanage) Append(orphan structs.Node) {
	o.Lock()
	defer o.Unlock()
	o.Orphans = append(o.Orphans, orphan)
	o.Len++
}

// Drop n items from the front of the Orphanage
// Lock is manually set from caller
func (o *Orphanage) DropN(n int) []structs.Node {
	droppedNodes := o.Orphans[:n]
	if n > len(o.Orphans) {
		log.Println("\x1b[31;1m"+"INDEX OUT OF RANGE ERROR!!!! LEN %d, WANTS TO DROP : %d"+"\x1b[0m", len(o.Orphans), n)
	}
	o.Orphans = o.Orphans[n:]
	o.Len -= uint32(n)
	return droppedNodes
}
