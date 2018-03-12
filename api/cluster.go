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

// Probably incomplete in its current state; extend as we see fit
// Used in Follow() to specify configuration details to Leader
type FollowerSettings struct {
	IPAddr    net.Addr
	TopicName string
}

// </DATA STRUCTURES>
///////////////////////////////////////////////////////////////////////////////////////////////////



/////////////////////////////////////////////////////////////////////////////////////////////////
// <API>

type Cluster interface {
	// Returns either the given nomination or a more suitable one
	// Suitability is determined by [TODO: Define heuristic]
	// Can return the following errors:
	// - DisconnectedError
	Vote(nomination Cluster) (node ClusterNode, err error)
	
	// Returns once all Followers have confirmed they accept the new Leader;
	// The Leader will then enter Leader Mode and other nodes will follow it
	// Can return the following errors:
	// - DisconnectedError
	Elect(nomination Cluster) (err error)

	// Confirm that followerID has been successfully written
	// These calls propagate to the Leader
	// Can return the following errors:
	// - DisconnectedError
	Confirm(followerID int64) (err error)
	
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

type ClusterNode struct {
	shared.Node
}

func (cn ClusterNode) Vote(nomination Cluster) (node Cluster, err error) {
	// TODO
}

func (cn ClusterNode) Elect(nomination Cluster) (err error) {
	// TODO
}

func (cn ClusterNode) Confirm(nomination Cluster) (err error) {
	// TODO
}

func (cn ClusterNode) Follow(settings FollowerSettings) (IPAddresses []net.Addr, err error) {
	// TODO
}

func (cn ClusterNode) FollowMe(IPAdrress net.Addr) (err error) {
	// TODO
}

func (cn ClusterNode) Connect(IPAddress net.Addr) (err error) {
	// TODO
}

// TODO: Separate Leader-specific calls so that Followers cannot invoke them
type Leader interface {
	ClusterNode

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

type LeaderNode struct {
	shared.Node
}

func (ln LeaderNode) Write(gpsCoordinates shared.GPSCoordinates) (err error) {
	// TODO
}

func (ln LeaderNode) Borrow(N uint32) (IPAddresses []net.Addr, err error) {
	// TODO
}

// </API>
///////////////////////////////////////////////////////////////////////////////////////////////////
