package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

// // Use Sync map for concurrent read and write
// // Use a write lock so we can't have data changed between a load and store
// var fm *sync.Map
// var writeLock *sync.Mutex

var TopicName string

// Write to VersionList
// For a Leader node, this list is guaranteed to be in order
// since a Leader serializes the Writes but we cannot guarantee
// the time arrival of those Writes to Follower nodes

// For a Follower node, we simply append the data, therefore needing a lock
// to ensure we aren't overwriting the array

var (
	VersionListLock      sync.RWMutex
	VersionList          []FileData
	LatestConfirmedIndex uint
) // Highest number that we know we have a write to

// FileSystem related errors //////
type FileSystemError string

func (e FileSystemError) Error() string {
	return fmt.Sprintf(string(e))
}

////////////////////////////////////

func MountFiles() error {
	// fm = new(sync.Map)
	// writeLock = &sync.Mutex{}
	// TODO: Make fault tolerant by reading files from drive

	VersionListLock = sync.RWMutex{}
	VersionList = make([]FileData, 0)
	fname := ""
	readFromDisk(fname)
	return nil
}

func WriteFile(topic, data string, version uint) error {
	if TopicName != "" && topic != TopicName {
		return errors.New("Writing to wrong topic")
	}

	TopicName = topic
	VersionList = append(VersionList, FileData{
		Version: version,
		Data:    data,
	})

	// writeLock.Lock()
	// defer writeLock.Unlock()

	// var newTopic = []string{data}

	//TODO: Write to Drive as well when done
	if err := writeToDisk(); err != nil {
		log.Println("ERROR WRITING TO DISK IN WRITEFILE")
		return err
	}
	return nil
}

// Returns confirmed writes
func ReadFile(topic string) ([]string, error) {
	topicData := make([]string, 0)
	if NodeMode == Leader && LatestConfirmedIndex == uint(len(VersionList)) {
		for _, fdata := range VersionList {
			topicData = append(topicData, fdata.Data)
		}

		return topicData, nil
	}

	if NodeMode == Follower {
		// TODO, Do we read from file for follower nodes
		return topicData, errors.New("It's a Follower, why you readnig from file")
	}

	return topicData, FileSystemError("General file error")
}

///////////////Writing to disk helpers /////////////////

func writeToDisk() error {
	fileData := ClusterData{
		Topic:   TopicName,
		Dataset: VersionList,
	}

	contents, err := json.MarshalIndent(fileData, "", "  ")
	fname := ""
	if err = ioutil.WriteFile(fname, contents, 0644); err != nil {
		log.Println("ERROR WRITING TO DISK")
		return err
	}

	return nil
}

func readFromDisk(fname string) (ClusterData, error) {
	contents, err := ioutil.ReadFile(fname)

	var clusterData ClusterData
	if err != nil {
		return clusterData, err
	}

	err = json.Unmarshal(contents, &clusterData)
	return clusterData, err
}

////////////End Writing to disk helpers /////////////////

/////////////// VersionList Helpers ///////////////////

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

/////////////// End VersionList Helpers ///////////////////
