package node

import (
	"../../structs"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type TODO struct{}
type Peer struct {
	Chan chan string
}

const HBTIMEOUT = 4
const HBINTERVAL = 2

var ServerClient *rpc.Client
var MinConnections uint8
var HBInterval uint32
var Peers sync.Map

func ConnectToServer(ip string) {
	LocalAddr, _ := net.ResolveTCPAddr("tcp", ":0")
	ServerAddr, _ := net.ResolveTCPAddr("tcp", ip)
	conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	if err != nil {
		fmt.Println("Could not connect to server")
	} else {
		fmt.Println("Connecting to server on:", conn.LocalAddr().String())
		ServerClient = rpc.NewClient(conn)
	}
}

// Not sure if this belongs here
func ServerRegister(addr string) {
	var resp structs.NodeSettings
	err := ServerClient.Call("TServer.Register", addr, &resp)
	if err != nil {
		fmt.Printf("Error in heartbeat::Register()\n%s", err)
	}
	MinConnections = resp.MinNumNodeConnections
	HBInterval = resp.HeartBeat
}

func ServerHeartBeat(addr string) {
	var _ignored bool
	fmt.Printf("starting server hb of: %d\n", HBInterval)
	interval := time.Duration(HBInterval / 2)
	heartbeat := time.Tick(interval * time.Millisecond)
	for {
		select {
		case <-heartbeat:
			ServerClient.Call("TServer.HeartBeat", addr, &_ignored)
		}
	}
}

// Heartbeat RPC for letting this node know that other guy is alive still
func (s *TODO) Heartbeat(id *string, reply *string) error {
	*reply = "ok"

	// Check if peer is in map
	peer, ok := getPeer(*id)
	if !ok {
		*reply = "Disconnected error"
		return fmt.Errorf("%s", *reply)
	}

	// Write to channel. Channel should not be nil.
	if peer.Chan != nil {
		peer.Chan <- "hb"
	} else {
		*reply = "CRITICAL ERROR - CHANNEL DNE"
		return fmt.Errorf("%s", *reply)
	}

	return nil
}

// Handles periodic sending of heartbeats to a single peer. The RPC connection
// should already be established, and the peer's channel should already be put
// into the Peers map structure.
func peerHbSender(id string, peerConn *rpc.Client) {
	peer, ok := getPeer(id)
	if !ok {
		return
	}

	ch := peer.Chan
	if ch == nil {
		return
	}

	// Note, could use a ticker, but not sure what behaviour would be if
	// tick occurs while not receiving, eg. occurs in the call.Done branch
	// before it starts waiting on timeout to complete
	timeout := createTimeout(HBINTERVAL)

	for true {
		arg := id
		var reply string

		call := peerConn.Go("Peer.Heartbeat", &arg, &reply, nil)
		if call == nil {
			// connection is dead - error
			ch <- "die"
			return
		}

		select {
		case <-timeout:
			// timeout occurs before call returns - error
			ch <- "die"
			return
		case <-call.Done:
			if call.Error != nil {
				ch <- "die"
				return
			}

			// Wait until timeout is done so that full interval has passed
			// before sending again.
			<-timeout
			timeout = createTimeout(HBINTERVAL)
		}
	}
}

// Handles heartbeat timeout checking for peers. Note that there is a check for
// both sending and receiving heartbeats from a peer. Seems unlikely, but there
// could be a case where a peer is taking heartbeats just fine, but is not
// sending any back.
func peerHbHandler(id string) {
	// Sanity checks - shouldn't ever happen
	peer, ok := getPeer(id)
	if !ok {
		return
	}
	ch := peer.Chan
	if ch == nil {
		return
	}

	// Cleanup routine for this long running function
	defer func() {
		// Delete peer from the map - can't talk to this guy anymore.
		Peers.Delete(id)
	}()

	// main loop of the heartbeat checking
	for true {
		timeout := createTimeout(HBTIMEOUT)

		select {
		case <-timeout:
			return // peer failure
		case msg := <-ch:
			switch msg {
			case "die":
				return // peer failure detected by another method
			case "hb":
				continue
			}
		}
	}
}

// Retrieve a peer from sync.Map - does the checking
func getPeer(id string) (peer *Peer, ok bool) {
	// Check if peer in the map and do type assertion
	val, ok := Peers.Load(id)
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

// Starts a goroutine that will write to the returned channel in <secs> seconds.
func createTimeout(secs time.Duration) chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(secs * time.Second)
		timeout <- true
	}()

	return timeout
}
