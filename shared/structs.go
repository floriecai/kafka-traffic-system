package shared

import (
	"net"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// DATA STRUCTURES
///////////////////////////////////////////////////////////////////////////////////////////////////

type GPSCoordinates struct {
	Lat float32
	Lon float32
}

type Type int

const (
	TypeServer = iota
	TypeProducer
	TypeConsumer
	TypeLeader
	TypeFollower
)

type NodeInfo struct {
	Address net.Addr
}

type NodeSettings struct {
	MinNumNodeConnections uint8
	HeartBeat             uint32
}

type Node struct {
	Type        Type
	Addr        net.Addr
	Settings    NodeSettings
	Coordinates GPSCoordinates
	Peers       []Node
}

type Topic struct {
	TopicName      string
	MinReplicasNum uint32
	Leaders        []Node
	Followers      []Node
}
