package node

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

type Mode int

type Peer struct {
	HbChan   chan string
	PeerConn *rpc.Client
	DeathFn  func(string)
}

const LEADER_ID = "leader"

const (
	Follower Mode = iota
	Leader
)

var NodeMode Mode = Follower

var PeerMap *sync.Map

var DirectFollowersList map[string]int // ip -> followerID
var FollowerListLock sync.Lock

var LeaderConn *rpc.Client

var FollowerId int = 1
var IdIncrementerLock sync.Lock


func BecomeLeader(ips []string, LeaderAddr string) (err error) {
	NodeMode = Leader

	for _, ip := range ips {
		LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
		if err != nil {
			continue
		}

		PeerAddr, err := net.ResolveTCPAddr("tcp", ip)
		if err != nil {
			continue
		}

		conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
		if err != nil {
			continue
		}

		client := rpc.NewClient(conn)

		var _ignored string
		fmt.Printf("Telling node with ip %s to follow me\n", ip)
		err = client.Call("Peer.FollowMe", LeaderAddr, &_ignored)
		if err != nil {
			continue
		}

		addPeer(ip, client, NodeDeathHandler)
	}
	return err
}

func FollowLeader(LeaderIp string) (err error) {
	DirectFollowersList = make(map[string]bool)

	LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return err
	}

	PeerAddr, err := net.ResolveTCPAddr("tcp", LeaderIp)
	if err != nil {
		return err
	}

	fmt.Printf("I am following the leader with ip %s now\n", LeaderIp)

	conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
	if err != nil {
		return err
	}

	// check if there is already a leader connection; if so, kill it.
	oldLeader, ok := getPeer(LEADER_ID)
	if ok {
		oldLeader.HbChan <- "die"
	}

	LeaderConn = rpc.NewClient(conn)
	addPeer(LEADER_ID, LeaderConn, NodeDeathHandler)

	return err
}

func ModifyFollowerList(ips []string, add bool) error {
	if add {
		for _, ip := range ips {
			if DirectFollowersList[ip] {
				log.Println(errors.New("Clustering: Follower is already known"))
			} else {
				DirectFollowersList[ip] = true
			}
		}
	} else {
		for _, ip := range ips {
			if !DirectFollowersList[ip] {
				log.Println(errors.New("Clustering: Follower is not known. Cannot remove follower"))
			} else {
				delete(DirectFollowersList, ip)
			}
		}
	}

	return nil
}

// Code from https://gist.github.com/jniltinho/9787946
func GeneratePublicIP() string {
	addrs, err := net.InterfaceAddrs()
	checkError(err, "GeneratePublicIP")

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String() + ":"
			}
		}
	}

	return "Could not find IP"
}

func checkError(err error, parent string) bool {
	if err != nil {
		fmt.Println(parent, ":: found error! ", err)
		return true
	}
	return false
}

// Adds a peer to the map and starts a heartbeat checking procedure for it.
func addPeer(id string, peerConn *rpc.Client, deathFn func(string)) {
	newPeer := Peer{make(chan string, 8), peerConn, deathFn}
	PeerMap.Store(id, newPeer)

	go peerHbSender(id)
	go peerHbHandler(id)
}

// Retrieve a peer from sync.Map - does the checking. ok will say whether
// or not a peer was successfully retrieved.
func getPeer(id string) (peer *Peer, ok bool) {
	// Check if peer in the map and do type assertion
	val, ok := PeerMap.Load(id)
	if !ok {
		return nil, false
	}
	p, ok := val.(Peer)
	if !ok {
		fmt.Println("CRITICAL ERROR: TYPE ASSERTION FAILED")
		return nil, false
	}

	return &p, true
}

func NodeDeathHandler(ip string) {
	// This is the death function in the case that this peer
	// dies. There will be more functionality added to this
	// later for sure. Maybe put into separate function.
	fmt.Printf("Oh no, %s died!\n", ip)
	switch(NodeMode) {
	case Follower:
		if ip == LEADER_ID {
			//TODO initiate consensus protocol
			fmt.Println("The leader has died, initiating consensus protocol")
		} else {
			//TODO react to death of other peers
			fmt.Println("Other peer has died, need to notify leader >>TODO<<")
		}	

	case Leader:
		fmt.Println("A node has died, need to remove it from everyone's follower list")
		// TODO
	default:
		// no default behavior
		fmt.Println("serious error occured in NodeDeathHandler")
	}
}

