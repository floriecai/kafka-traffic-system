package producer

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
	return fmt.Sprintf("Producer: Cannot connect to [%s]", string(e))
}

// Could not guarantee Write to be replicated the minimum number of times
type InsufficientReplicasError string

func (e InsufficientReplicasError) Error() string {
	return fmt.Sprintf("Producer: Could not replicate minimum number of times")
}

// </ERROR DEFINITIONS>
///////////////////////////////////////////////////////////////////////////////////////////////////



/////////////////////////////////////////////////////////////////////////////////////////////////
// <API>

type Producer interface {
	// Returns the Topic on which the client operates
	// Can return the following errors:
	// - DisconnectedError
	// - DuplicateTopicNameError
	CreateTopic(topicName string) (topic shared.Topic, err error)
	
	// Returns the cluster with the given TopicName
	// Can return the following errors:
	// - TopicClusterDoesNotExistError
	// - DisconnectedError
	GetTopic(topicName string) (topic shared.Topic, err error)
	
	// Flooding procedure initiated by the Leader to propagate a Write
	// Can return the following errors:
	// - InsufficientReplicasError
	Write(topicName string, gpsCoordinates shared.GPSCoordinates) (err error)
}

// </API>
///////////////////////////////////////////////////////////////////////////////////////////////////
