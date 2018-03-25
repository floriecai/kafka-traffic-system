package node


import ("sync"
		"net/rpc"
		"net"
)
type Mode int

const (
	Follower Mode = iota
	Leader
)

var NodeMode Mode = Follower

var PeerMap *sync.Map

var DirectFollowersList []string

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
	DirectFollowersList = msg.FollowerIps

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