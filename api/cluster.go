package cluster

import (
	"fmt"
	"net"

	"../shared"
)



///////////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// Cannot connect
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("Cluster: Cannot connect to [%s]", string(e))
}

// Not enough Followers to give
type InsufficientNodesError string

func (e InsufficientNodesError) Error() string {
	return fmt.Sprintf("Cluster: Not enough followers to give")
}

// </ERROR DEFINITIONS>
///////////////////////////////////////////////////////////////////////////////////////////////////



///////////////////////////////////////////////////////////////////////////////////////////////////
// <DATA STRUCTURES>

// Used in Follow() to specify 
type FollowerSettings struct {
	// TODO
	IPAddr    net.Addr
	TopicName string
}

// </DATA STRUCTURES>
///////////////////////////////////////////////////////////////////////////////////////////////////



/////////////////////////////////////////////////////////////////////////////////////////////////
// <API>

type ClusterNode interface {
	// Returns either the given nomination or a more suitable one
	// Suitability is determined by [TODO: Define heuristic]
	// Can return the following errors:
	// - DisconnectedError
	Vote(nomination ClusterNode) (node ClusterNode, err error)
	
	// Returns once all Followers have confirmed they accept the new Leader;
	// The Leader will then enter Leader Mode and other nodes will follow it
	// Can return the following errors:
	// - DisconnectedError
	Elect(nomination ClusterNode) (err error)

	// Confirm that followerID has been successfully written
	// These calls propagate to the Leader
	// Can return the following errors:
	// - DisconnectedError
	Confirm(followerID int64) ()
	
	// Returns a list of IP addresses to which a Follower may connect
	// Can return the following errors:
	// - DisconnectedError
	Follow(settings FollowerSettings) (IPAddresses []net.Addr, err error)
	
	// Follow the given IP address for cluster creation or node migration
	// Can return the following errors:
	// - DisconnectedError
	FollowMe(IPAdrress net.Addr) (err error)
	
	// Connect to ClusterNode at the given IP address
	// Can return the following errors:
	// - DisconnectedError
	Connect(IPAddress net.Addr) (err error)
}

// TODO: Separate Leader-specific calls so that Followers cannot invoke them
type LeaderNode interface {
	// Flooding procedure initiated by the Leader to propagate a Write
	// Can return the following errors:
	// - DisconnectedError
	Write(gpsCoordinates shared.GPSCoordinates) (err error)
	
	// Returns N Followers subject to node migration
	// Can return the following errors:
	// - DisconnectedError
	// - InsufficientNodesError
	Borrow(N uint32) (IPAddresses []net.Addr, err error)
}

// </API>
///////////////////////////////////////////////////////////////////////////////////////////////////
