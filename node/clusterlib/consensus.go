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
)

// Maximum number of seconds for consensus job to wait for before timeout error.
const DATUM_CONSENSUS_TIMEOUT = 10

// Number of seconds to wait until election is considered complete
const ELECTION_COMPLETE_TIMEOUT = 5

// Number of seconds to wait after election complete to let results come in
const ELECTION_WAIT_FOR_RESULTS = 10

// TODO: replace this with proper types in the proper places
type TODO struct{}

var dataChannels sync.Map

// Election related file-global vars
var electionInProgress bool = false
var electionNumRequired int
var electionNumAccepted int
var electionLock sync.Mutex
var electionComplete bool = false
var electionCompleteCond sync.Cond
var chosenLeaderId string
var chosenLeaderLatestNum int
var nominationCompleteCh chan bool = make(chan bool, 1)

// Function WriteAfterConsent will start a consensus job, which will write to
// disk and propagate the confirmation of the data for the datumNum provided if
// consensus is reached.
func WriteAfterConsent(numFollowers int, datum TODO, datumNum int) {
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

		electionCompleteCond.L.Lock()
		electionComplete = true
		electionCompleteCond.L.Unlock()

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

// Function CheckIfAcceptPeer should be called when a peer self-nomination is
// received. This function will return true if it consents for that peer to be
// the leader, and false otherwise
func CheckPeerNominateAccept(peerLatestNum int, peerId string) bool {
	electionLock.Lock()

	// reject immediately if no good
	// wait for accept if peerLatestNum is greatest - there could be
	// another one yet to arrive that is greater.
	// - condvar, and select on timeout
	if peerLatestNum < chosenLeaderLatestNum {
		electionLock.Unlock()
		return false
	} else if peerLatestNum == chosenLeaderLatestNum {
		result := tieBreakLatestNum(peerId)
		if !result {
			electionLock.Unlock()
			return false
		}
	}

	// If reached here, the peerId is now the chosen leader for now.
	chosenLeaderLatestNum = peerLatestNum
	chosenLeaderId = peerId
	electionLock.Unlock()

	electionCompleteCond.L.Lock()
	if !electionComplete {
		electionCompleteCond.Wait()
	}
	electionCompleteCond.L.Unlock()

	// check if chosenLeaderId is still peerId, and return the result
	return (peerId == chosenLeaderId)
}

// Function PeerAcceptThisNode should be called if a peer has accepted that
// this node should be leader. Currently is not responsible for ensuring that
// each peer has only sent their acceptance once. (maybe it should be though)
func PeerAcceptThisNode() {
	electionLock.Lock()
	defer electionLock.Unlock()

	electionNumAccepted++
	if electionNumAccepted > electionNumRequired {
		select {
		case nominationCompleteCh <- true:
			fmt.Println("Writed completion to nominationComplete")
		default:
			fmt.Println("nominationCompleteCh in full")
		}
	}
}

// This function starts a consensus job. A caller should update the job when
// new messages are received using the update channel. The function parameter
// fn will be called if there is a consensus reached. Consensus is considered
// successful when there are a number of writes to the updateChannel that is
// at least half of numFollowers.
func createConsensusJob(numFollowers int, timeout time.Duration, fn func(), datumNum int) (updateChannel chan<- bool) {
	uCh := make(chan bool, 32)

	timeoutCh := createTimeout(timeout)

	go func() {
		// Function will keep track of number accepted and rejected
		numAccepted := 0
		numRejected := 0
		done := (numFollowers / 2) + (numFollowers % 2)

		for {
			select {
			case accepted := <-uCh:
				if accepted {
					numAccepted += 1
					if numAccepted >= done {
						fn()
					}
				} else {
					numRejected += 1
					if numRejected >= done {
						// Currently does nothing if a datum
						// is rejected. It should probably do
						// something though, like ask if its
						// peers think it is the leader.
						fmt.Printf("DATUM %d rejected\n", datumNum)
					}
				}
			case <-timeoutCh:
				// Delete the channel from the map, but do not
				// close (unsafe to do so). I trust the golang
				// GC to clean this up once it's not in the map.
				dataChannels.Delete(datumNum)

				// safe to close the timeout channel
				close(timeoutCh)
				return
			}
		}
	}()

	return uCh
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
