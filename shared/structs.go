package shared

import (
	"net"
)

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
