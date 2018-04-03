package node

import (
	"errors"
	"fmt"
	"sync"
)

type FileData struct {
	Version uint   `json:"version"`
	Data    string `json:"data"`
}

type ClusterData struct {
	Topic   string     `json:"topic"`
	Dataset []FileData `json: "dataset"`
}

// Use Sync map for concurrent read and write
// Use a write lock so we can't have data changed between a load and store
var fm *sync.Map
var writeLock *sync.Mutex

var TopicName string

// Write to VersionList
var VersionList = make([]FileData, 0)

// Highest number that we know we have a write to
var LatestConfirmedIndex uint

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

func WriteFile(topic, data string, version uint) error {
	if TopicName != "" && topic != TopicName {
		return errors.New("Writing to wrong topic")
	}

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

// VersionList Helpers

func HasCompleteList() bool {
	for i, fdata := range VersionList[LatestConfirmedIndex:] {
		offset := uint(i) + LatestConfirmedIndex
		if offset != fdata.Version {
			LatestConfirmedIndex = offset - 1
			return false
		}
	}
	return true
}

func GetConfirmedWrites() []FileData {
	confirmedWrites := make([]FileData, 0)
	for i := 0; uint(i) < LatestConfirmedIndex; i++ {
		confirmedWrites = append(confirmedWrites, VersionList[i])
	}

	return confirmedWrites
}
