package structs

import (
	//"net"
	"net/rpc"
)

type NodeSettings struct {
	MinReplicas uint8  `json:"min-replicas"`
	HeartBeat   uint32 `json:"heartbeat"`
	ClusterSize uint8  `json:"cluster-size"`
}

type Node struct {
	Address         string
	Client          *rpc.Client
	RecentHeartbeat int64
	IsLeader        bool
}

type Topic struct {
	TopicName   string
	MinReplicas uint8
	Leaders     []string // list of ip addresses of each node
	Followers   []string // list of ip addresses of each node
}

////////////////////// RPC STRUCTS //////////////////////

type AdditionalNodeRequest struct { // nothing for now
}

type ExtraNodeResponse struct {
	NewNodeIpAddr string
}

/////////////////// RPC STRUCTS END ////////////////////
