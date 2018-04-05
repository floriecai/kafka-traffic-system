package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
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

// DataPath where files are written to disk
var DataPath string

var TopicName string

// For a Leader node, this list is guaranteed to be in order and continuous
// since a Leader serializes the Writes but we cannot guarantee
// the time arrival of those Writes to Follower nodes
var (
	VersionListLock sync.Mutex
	VersionList     []FileData
	IsSorted        bool // TODO: To avoid doing multiple sorts
)

// FileSystem related errors //////
type FileSystemError string

func (e FileSystemError) Error() string {
	return fmt.Sprintf(string(e))
}

type InsufficientConfirmedWritesError string

func (e InsufficientConfirmedWritesError) Error() string {
	return fmt.Sprintf(string(e))
}

type IncompleteDataError string

func (e IncompleteDataError) Error() string {
	return fmt.Sprintf("There is an incomplete dataset. Cannot read.")
}

////////////////////////////////////

func MountFiles(path string) {
	VersionListLock = sync.Mutex{}
	VersionList = make([]FileData, 0)
	fname := filepath.Join(path, "data.json")

	// First time a node has registered with a server
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		_, err := os.Create(fname)
		if err != nil {
			checkError(err, "MountFiles CreateFiles")
			log.Fatalf("Couldn't create file [%s]", fname)
		}

		return
	}

	// Rejoining node, read topic data from disk
	var clusterData ClusterData
	err := readFromDisk(fname, &clusterData)
	TopicName = clusterData.Topic
	VersionList = clusterData.Dataset

	if err != nil {
		checkError(err, "MountFiles DiskRead")
		log.Fatalf("Could not read from disk")
	}
}

// Add the Write to the VersionList and commit Write to disk
func WriteNode(topic, data string, version uint) error {
	if TopicName != "" && topic != TopicName {
		return errors.New("Writing to wrong topic")
	}

	TopicName = topic
	VersionList = append(VersionList, FileData{
		Version: version,
		Data:    data,
	})

	if err := writeToDisk(DataPath); err != nil {
		log.Println("ERROR WRITING TO DISK IN WRITEFILE")
		return err
	}
	return nil
}

// Returns confirmed writes the node contains
// Errors:
// IncompleteDataError - Not all writes have been received
func ReadNode(topic string) ([]string, error) {
	VersionListLock.Lock()
	defer VersionListLock.Unlock()
	confirmedWrites := GetConfirmedWrites()

	if len(confirmedWrites) != len(VersionList) {
		return confirmedWrites, IncompleteDataError("")
	}

	return confirmedWrites, nil
}

///////////////Writing to disk helpers /////////////////
func writeToDisk(path string) error {
	fileData := ClusterData{
		Topic:   TopicName,
		Dataset: VersionList,
	}

	contents, err := json.MarshalIndent(fileData, "", "  ")
	if err = ioutil.WriteFile(path, contents, 0644); err != nil {
		log.Println("ERROR WRITING TO DISK")
		return err
	}

	return nil
}

func readFromDisk(fname string, clusterData *ClusterData) error {
	contents, err := ioutil.ReadFile(fname)
	if err != nil {
		checkError(err, "readFromDisk")
		return err
	}

	if len(contents) == 0 {
		log.Println("Reading from disk ... No ClusterData")
		return nil
	}

	err = json.Unmarshal(contents, clusterData)
	return err
}

////////////End Writing to disk helpers /////////////////

/////////////// VersionList Helpers ///////////////////

// Sorts VersionList by its version number and returns the first index
// that does not match its version number
// Returns -1 if all indices match its version number
func sortVersionList() int {
	sort.Slice(VersionList, func(i, j int) bool {
		return VersionList[i].Version > VersionList[j].Version
	})

	for i, fdata := range VersionList {
		if uint(i) != fdata.Version {
			return i
		}
	}

	return -1
}

// Returns the longest list of continuous ordered writes
// VersionList: [1,2,3,4,6,7,8] will return [1,2,3,4]
// Note: The caller of this function is responsible for locking
//       since it likely has to do additional work related to the VersionList
func GetConfirmedWrites() []string {
	firstMismatch := sortVersionList()
	if firstMismatch == -1 {
		firstMismatch = len(VersionList)
	}

	writes := make([]string, 0)
	for _, fdata := range VersionList[:firstMismatch] {
		writes = append(writes, fdata.Data)
	}
	return writes
}

/////////////// End VersionList Helpers ///////////////////
