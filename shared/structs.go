package shared

import (
	"net"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// <DATA STRUCTURES>

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

type Node struct {
	Type        Type
	Addr        net.Addr
	Coordinates GPSCoordinates
	Peers       []Node
}

type Topic struct {
	TopicName      string
	MinReplicasNum uint32
	Leaders        []Node
	Followers      []Node
}

// </DATA STRUCTURES>
/////////////////////////////////////////////////////////////////////////////////////////////////
