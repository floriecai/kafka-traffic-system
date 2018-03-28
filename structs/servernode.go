package structs

import (
	"net"
	"net/rpc"
)

type NodeSettings struct {
	MinNumNodeConnections uint8  `json:"min-num-node-connections"`
	HeartBeat             uint32 `json:"heartbeat"`
}

type Node struct {
	Address         net.Addr
	Client          *rpc.Client
	RecentHeartbeat int64
	IsLeader        bool
}

type Topic struct {
	TopicName   string
	MinReplicas uint8
	Leaders     []Node
	Followers   []Node
}

////////////////////// RPC STRUCTS //////////////////////

type AdditionalNodeRequest struct { // nothing for now
}

type ExtraNodeResponse struct {
	NewNodeIpAddr string
}

/////////////////// RPC STRUCTS END ////////////////////
