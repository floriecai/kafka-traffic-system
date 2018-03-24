package node


import ("sync"
		
)
type Mode int

const (
	Follower Mode = iota
	Leader
)

var NodeMode Mode = Follower

var PeerMap *sync.Map


func BecomeLeader(ips []string) error {
	NodeMode = Leader

	for ip := range ips {
		
	}

}