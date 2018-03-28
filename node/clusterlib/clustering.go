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

const (
	Follower Mode = iota
	Leader
)

var NodeMode Mode = Follower

var PeerMap *sync.Map

var DirectFollowersList map[string]bool // ip -> true

var LeaderConn *rpc.Client

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
		err = client.Call("Peer.FollowMe", LeaderAddr, &_ignored)
		if err != nil {
			continue
		}

		PeerMap.Store(ip, client)

	}
	return err
}

func FollowLeader(msg FollowMeMsg) (err error) {
	DirectFollowersList = make(map[string]bool)

	for _, ip := range msg.FollowerIps {
		DirectFollowersList[ip] = true
	}

	LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return err
	}

	PeerAddr, err := net.ResolveTCPAddr("tcp", msg.LeaderIp)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
	if err != nil {
		return err
	}

	LeaderConn = rpc.NewClient(conn)
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
