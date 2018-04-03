package node

import (
	"fmt"
	"net"
	"net/rpc"
	"time"

	"../../structs"
)

const HBTIMEOUT = 4
const HBINTERVAL = 2

var ServerClient *rpc.Client
var MinConnections uint8
var HBInterval uint32

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

// Logic of the heartbeat function
func PeerHeartbeat(id string, reply *string) error {
	*reply = "ok"

	// Check if peer is in map, then write to its heartbeat channel
	peer, ok := PeerMap.Get(id)
	if !ok {
		*reply = "Disconnected error"
		return fmt.Errorf("%s", *reply)
	}
	peer.HbChan <- "hb"

	return nil
}

// Handles periodic sending of heartbeats to a single peer. The RPC connection
// should already be established, and the peer's channel should already be put
// into the PeerMap structure.
func peerHbSender(id string) {
	peer, ok := PeerMap.Get(id)
	if !ok {
		return
	}

	// Note, could use a ticker, but not sure what behaviour would be if
	// tick occurs while not receiving, eg. occurs in the call.Done branch
	// before it starts waiting on timeout to complete
	timeout := createPeerTimeout(HBINTERVAL)

	for true {
		arg := id
		var reply string

		call := peer.PeerConn.Go("Peer.Heartbeat", arg, &reply, nil)
		if call == nil {
			// connection is dead - error
			peer.HbChan <- "die"
			return
		}

		select {
		case <-timeout:
			// timeout occurs before call returns - error
			peer.HbChan <- "die"
			return
		case <-call.Done:
			if call.Error != nil {
				fmt.Printf("Peer.Heartbeat error: %s\n", reply)
				peer.HbChan <- "die"
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
	peer, ok := PeerMap.Get(id)
	if !ok {
		return
	}

	// Cleanup routine for this long running function. There is more than
	// one exit point in the loop  so defer this exit function. This is the
	// single point of deletion for a peer connection.
	defer func() {
		// Delete peer from the map - can't talk to this guy anymore.
		PeerMap.Delete(id)
		peer.PeerConn.Close()

		fmt.Printf("Peer %s connection has died!", id)
		peer.DeathFn(id)
	}()

	// Heartbeat checking loop - does not exit until a peer disconnects
	for true {
		timeout := createTimeout(HBTIMEOUT)

		select {
		case <-timeout:
			// Peer failure detected!
			return
		case msg := <-peer.HbChan:
			switch msg {
			case "die":
				// Peer failure by another function, exit
				return
			case "hb":
				continue
			}
		}
	}
}

// Starts a goroutine that will write to the returned channel in <secs> seconds.
func createPeerTimeout(secs time.Duration) (timeout chan bool) {
	timeout = make(chan bool, 1)
	go func() {
		time.Sleep(secs * time.Second)
		timeout <- true
	}()
	return timeout
}
