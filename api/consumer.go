package consumer

import (
	"fmt"

	"../shared"
)



///////////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// There is at least one successful Write operation to
// this Cluster that could not be retrieved
type DataUnvailableError string

func (e DataUnvailableError) Error() string {
	return fmt.Sprintf("Consumer: Unavailable data")
}

// Cannot connect to server
type DisconnectedServerError string

func (e DisconnectedServerError) Error() string {
	return fmt.Sprintf("Consumer: Cannot connect to server on [%s]", string(e))
}

// No Cluster with the given TopicName exists
type TopicClusterDoesNotExistError string

func (e TopicClusterDoesNotExistError) Error() string {
	return fmt.Sprintf("Consumer: Topic with name [%s] does not exist", string(e))
}

// </ERROR DEFINITIONS>
///////////////////////////////////////////////////////////////////////////////////////////////////



/////////////////////////////////////////////////////////////////////////////////////////////////
// <API>

type Consumer interface {
	// Returns the Topic associated with the given gpsCoordinates
	// Can return the following errors:
	// - TopicClusterDoesNotExistError
	GetCurrentLocationCluster(gpsCoordinates GPSCoordinates) (topic shared.Topic, err error)

	// Returns the Topic the client will connect to and read from
	// Can return the following errors:
	// - TopicClusterDoesNotExistError
	// - DisconnectedServerError
	GetCluster(topicName string) (topic shared.Topic, err error)

	// Returns a list of all GPSCoordinates that have been written to the Topic
	// Can returen the following errors:
	// - DataUnvailableError
	Read(topicName string) (gpsCoordinates shared.GPSCoordinates, err error)
}

// </API>
///////////////////////////////////////////////////////////////////////////////////////////////////
