package node

import (
	"fmt"
	"sync"
)

// Use Sync map for concurrent read and write
// Use a write lock so we can't have data changed between a load and store
var fm *sync.Map
var writeLock *sync.Mutex

// key: Topic -- value: []string where the strings are data

// FileSystem related errors //////
type FileSystemError struct {
	Reason string
}

func (e FileSystemError) Error() string {
	return fmt.Sprintf(e.Reason)
}

////////////////////////////////////

func MountFiles() error {
	fm = new(sync.Map)
	writeLock = &sync.Mutex{}
	// TODO: Make fault tolerant by reading files from drive
	return nil
}

func WriteFile(topic, data string) error {
	var topicData []string

	writeLock.Lock()
	defer writeLock.Unlock()

	var newTopic = []string{data}

	// Add the data if it's a new topic
	rawData, exists := fm.LoadOrStore(topic, newTopic)
	topicData = rawData.([]string)

	// There was already existing data for this, append
	if exists {
		topicData = append(topicData, data)
		fm.Store(topic, topicData)
	}

	//TODO: Write to Drive as well when done
	return nil
}

func ReadFile(topic string) ([]string, error) {
	var topicData []string

	rawData, ok := fm.Load(topic)
	if ok {
		// Need to type assert
		topicData = rawData.([]string)
		return topicData, nil
	} else {
		return topicData, FileSystemError{"a general error"}
	}
}
