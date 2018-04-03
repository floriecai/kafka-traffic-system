/*

This file contains the consensus protocol functions. This is both for data
consensus as well as leader nomination consensus.

Current TODOs:
- peerId: what type is this?
- tieBreakLatestNum: not sure about implementation (depends on above)
- WriteAfterConsent: nothing currently done when datum consensus is achieved
- leader election: what do on completion?

*/
package node

import (
	"fmt"
	"sync"
	"time"
	"math"
	"net"
	"net/rpc"
)

// Maximum number of seconds for consensus job to wait for before timeout error.
const DATUM_CONSENSUS_TIMEOUT = 10

// Number of seconds to wait until election is considered complete
const ELECTION_COMPLETE_TIMEOUT = 5

// Number of seconds to wait after election complete to let results come in
const ELECTION_WAIT_FOR_RESULTS = 10

var dataChannels sync.Map

// Election related file-global vars
var electionInProgress bool = false
var electionNumRequired int
var electionNumAccepted int
var electionLock sync.Mutex
var electionComplete bool = false
var electionUpdateCond sync.Cond
var chosenLeaderId string
var chosenLeaderLatestNum int
var nominationCompleteCh chan bool = make(chan bool, 1)

var PotentialFollowerIps []string
var receiveFollowerChannel chan string
var MyAddr string

// Initial entry point to the consensus protocol
// Called when the leader heartbeat dstops
func StartConsensusProtocol() {

	lowestFollowerIp, lowestFollowerId := ScanFollowerList()

	// create Consensus job that returns an update channel and a channel for the func to receive followers on
	var updateChannel chan bool

	for len(DirectFollowersList) > 0 {
	// case 1: we are the lowest follower and likely to become leader
		if (FollowerId == lowestFollowerId) {
			fmt.Println("Expecting to become leader")

			electionLock.Lock()
			electionInProgress = true
			electionLock.Unlock()

			updateChannel, receiveFollowerChannel = StartElection()
			// block on update channel
			becameLeader := <- updateChannel
			// when channel returns and it's true then start become leader protocol
			// Assume PotentialFollowerIps was filled up

			electionLock.Lock()
			electionInProgress = false
			electionLock.Unlock()

			if becameLeader {

				BecomeLeader(PotentialFollowerIps, lowestFollowerIp)
				// TODO: Notify server once you've become leader
				break
			}
		} else {
	// case 2: we should connect to the lowest follower
			fmt.Println("Try to follow this leader:", lowestFollowerIp)
			updateChannel, receiveFollowerChannel = Nominate()

			nominationAccepted := <- updateChannel
			if nominationAccepted {
				break
			}
		}
	}

}

// Function WriteAfterConsent will start a consensus job, which will write to
// disk and propagate the confirmation of the data for the datumNum provided if
// consensus is reached.
/*
func WriteAfterConsent(numFollowers int, datumNum int) {
	fn := func() {
		fmt.Printf("Wrote datumNum %d\n", datumNum)
		//WriteToFile(datum)
		//PropagateSuccess(datum)
	}

	updateCh := createConsensusJob(numFollowers,
		time.Second*DATUM_CONSENSUS_TIMEOUT, fn, datumNum)

	dataChannels.Store(datumNum, updateCh)
}

// Function UpdateDatumConsent should be called from an RPC call by a follower
// stating their consent for a piece of data. It will return an error if the
// datum consent has  timed out or completed, or if somehow there is a type
// assertion error.
func UpdateDatumConsent(datumNum int, accepted bool) error {
	val, ok := dataChannels.Load(datumNum)
	if !ok {
		return fmt.Errorf("Consent job timed out or completed")
	}

	ch, ok := val.(chan<- bool)
	if !ok {
		errorStr := "CRITICAL ERROR - BAD TYPE IN DATA MAP"
		fmt.Printf("%s\n", errorStr)
		return fmt.Errorf("%s", errorStr)
	}

	ch <- accepted
	return nil
}

// Function StartElection should only be called if the node is currently not
// a leader. If it is the leader, then this function will do nothing. If there
// is an election already in progress, then it will be a no-op.
//
// It could be that this node received some more data after the election
// already started, but it had to receive that data from somewhere, so assume
// that there is some peer with a latest datum number greater or equal to that
// value if that is the case.
//
// There should be a timeout period between last data received and the recognized
// death of the current leader, which should provide sufficient time for data
// to be propagated fully between followers, so the above case described seems
// unlikely.
func StartElection(myLatestNum int, myId string, numAcceptRequired int) {
	electionLock.Lock()
	defer electionLock.Unlock()

	if electionInProgress {
		return
	}

	electionComplete = false
	electionInProgress = true
	electionNumAccepted = 0
	electionNumRequired = numAcceptRequired

	// create a new nomination channel so that it is guaranteed to be empty
	nominationCompleteCh = make(chan bool, 1)

	thisNodeId := myId
	chosenLeaderId = myId
	chosenLeaderLatestNum = myLatestNum

	go func() {
		// Give some time for propagations to occur and consensus to be
		// created.
		time.Sleep(ELECTION_COMPLETE_TIMEOUT)

		electionUpdateCond.L.Lock()
		electionComplete = true
		electionUpdateCond.L.Unlock()

		if thisNodeId == chosenLeaderId {
			// This sleep is to ensure that there is enough time after
			// completion for peer nodes to also finish and propagate their
			// results to peers.
			result_timeout := createTimeout(ELECTION_WAIT_FOR_RESULTS)

			// Don't actually do anything with these channels, just
			// wait until one of them is written to before continuing.
			select {
			case <-nominationCompleteCh:
				fmt.Println("Nomination completed!")
			case <-result_timeout:
				fmt.Println("Nomination process timed out")
			}

			if electionNumAccepted > electionNumRequired {
				fmt.Println("this node has won the election; WHAT DO")
			}
		}
	}()
}
*/
// Function Nominate is called when the peer decides there is a leader
// to join, it will return 2 channels, one for updating the status and one
// for receiving the update from peer rpc
func Nominate() (updateCh chan bool, receiveFollowerCh chan string) {
	updateCh = make(chan bool, 32)
	receiveFollowerCh = make(chan string, 32)
	timeoutCh := createTimeout(ELECTION_WAIT_FOR_RESULTS)


	go func() {
		for {
			select {
			// Receive a new FollowMe
			// end the election status
			case <- receiveFollowerCh:
				electionLock.Lock()
				/////////////
				updateCh <- true
				/////////////
				electionLock.Unlock()
				return

			case <- timeoutCh:
				// Delete the channel from the map, but do not
				// close (unsafe to do so). I trust the golang
				// GC to clean this up once it's not in the map.
				//dataChannels.Delete(datumNum)

				// safe to close the timeout channel
				close(timeoutCh)
				updateCh <- false
				return
			}
		}
	}()
	return updateCh, receiveFollowerCh
}

// Function PeerAcceptThisNode should be called if a peer has accepted that
// this node should be leader. Currently is not responsible for ensuring that
// each peer has only sent their acceptance once. (maybe it should be though)
func PeerAcceptThisNode(ip string) error {
	electionLock.Lock()
	defer electionLock.Unlock()

	if electionInProgress {
		receiveFollowerChannel <- ip
		return nil
	} else {		
		// from clustering.go
		// it's likely this cluster is trying to join after
		// an election so just accept it
		LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
		if err != nil {
			return err
		}

		PeerAddr, err := net.ResolveTCPAddr("tcp", ip)
		if err != nil {
			return err
		}

		conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
		if err != nil {
			return err
		}

		client := rpc.NewClient(conn)

		var _ignored string

		FollowerListLock.RLock()
		////////////////////////////
		// It's ok if it fails, gaps in follower ID sequence will not mean anything
		FollowerId += 1 
		msg := FollowMeMsg{MyAddr, DirectFollowersList, FollowerId}
		fmt.Printf("Telling node with ip %s to follow me\n", ip)
		err = client.Call("Peer.FollowMe", msg, &_ignored)
		////////////////////////////
		FollowerListLock.RUnlock()
		if err != nil {
			return err
		}

		// Write lock when modifying the direct followers list
		FollowerListLock.Lock()
		DirectFollowersList[ip] = FollowerId
		FollowerListLock.Unlock()
		// unlock

		addPeer(ip, client, NodeDeathHandler, FollowerId)
		return nil
	}

}

// This function starts a consensus job. A caller should update the job when
// new messages are received using the update channel. Consensus is considered
// successful when there are a number of writes to the potential follower list that is
// at least the min connections needed for a cluster
func StartElection() (updateCh chan bool, receiveFollowerCh chan string) {
	updateCh = make(chan bool, 32)
	receiveFollowerCh = make(chan string, 32)
	timeoutCh := createTimeout(ELECTION_WAIT_FOR_RESULTS)

	go func() {
		for {
			select {
			// Receive a new FollowerIp
			// Add it to potential followers list
			// if we have enough then we end the election process
			case follower := <-receiveFollowerCh:
				electionLock.Lock()
				/////////////
				PotentialFollowerIps = append(PotentialFollowerIps, follower)
				if len(PotentialFollowerIps) >= int(MinConnections) {
					updateCh <- true
					electionLock.Unlock()
					return
				}
				/////////////
				electionLock.Unlock()

			case <-timeoutCh:
				// Delete the channel from the map, but do not
				// close (unsafe to do so). I trust the golang
				// GC to clean this up once it's not in the map.
				//dataChannels.Delete(datumNum)

				// safe to close the timeout channel
				close(timeoutCh)
				updateCh <- false
				return
			}
		}
	}()

	return updateCh, receiveFollowerCh
}

// Function tieBreakLatestNum contains the logic to choose whether or not a
// peer ID is better than the current one
func tieBreakLatestNum(peerId string) bool {
	// TODO: what to do here
	return false
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

// return the lowest follower's ID and the corresponding IP
func ScanFollowerList() (lowestFollowerIp string, lowestFollowerId int) {
	FollowerListLock.Lock()
	defer FollowerListLock.Unlock()

	lowestFollowerId = math.MaxInt32
	lowestFollowerIp = ":0"

	// Scan for lowest follower ID
	for ip, id := range DirectFollowersList {
		if id < lowestFollowerId {
			lowestFollowerId = id
			lowestFollowerIp = ip
		}
	}
	delete(DirectFollowersList, lowestFollowerIp)
	return lowestFollowerIp,lowestFollowerId
}
