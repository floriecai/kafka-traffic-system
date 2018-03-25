package node

// Structs for node-based p2p messages

type FollowMeMsg struct {
	LeaderIp string
	FollowerIps []string
}