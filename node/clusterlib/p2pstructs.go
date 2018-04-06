package node

// Structs for node-based p2p messages

type FollowMeMsg struct {
	LeaderIp    string
	FollowerIps map[string]int
	YourId      int
}

type ModFollowerListMsg struct {
	FollowerIp string
	FollowerId int
}

type PropagateWriteReq struct {
	Topic      string
	Data       string
	VersionNum int
	LeaderId   string
}
