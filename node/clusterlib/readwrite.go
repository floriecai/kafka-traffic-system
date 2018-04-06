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

const GREEN_COL = "\x1b[32;1m"
const ERR_COL = "\x1b[31;1m"
const ERR_END = "\x1b[0m"

type FileData struct {
	Version int    `json:"version"`
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

	// Writes may come in different order so this index is to show
	// where in VersionList are the Writes no longer ordered
	// i.e. [1,2,3,5,6], the first mismatch is 3
	// since V5 does not match its index of 4. (note: WriteId's begin at 1)
	// If FirstMismatch == -1 or Len(VersionList), we have all the writes
	FirstMismatch int = -1
	isSorted      bool

	// Channel for passing the new WriteId back to node main package
	writeIdCh chan int
)

// FileSystem related errors //////
type FileSystemError string

func (e FileSystemError) Error() string {
	return fmt.Sprintf(string(e))
}

type InsufficientConfirmedWritesError string

func (e InsufficientConfirmedWritesError) Error() string {
	return fmt.Sprintf("Could not replicate write enough times. %s", string(e))
}

type IncompleteDataError string

func (e IncompleteDataError) Error() string {
	return fmt.Sprintf("There is an incomplete dataset. Cannot read.")
}

////////////////////////////////////

func MountFiles(path string, writeIdCh chan int) {
	VersionListLock = sync.Mutex{}
	VersionList = make([]FileData, 0)
	writeIdCh = writeIdCh
	DataPath = path
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
func WriteNode(topic, data string, version int) error {
	if TopicName != "" && topic != TopicName {
		return errors.New("Writing to wrong topic")
	}

	TopicName = topic
	VersionListLock.Lock()

	// Minor optimizations.
	// We're assuming that writes often come in order and if its greater than the last
	// item in the list, just append it.
	// We keep track of isSorted to avoid having to iterate through the list
	// when calling GetConfirmedReads
	versionLen := len(VersionList)
	if versionLen == 0 {
		isSorted = true
	} else {
		last := VersionList[versionLen-1]
		if last.Version > version {
			isSorted = false
		}
	}

	VersionList = append(VersionList, FileData{
		Version: version,
		Data:    data,
	})
	VersionListLock.Unlock()

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
	confirmedWrites := GetConfirmedWrites()

	if len(confirmedWrites) != len(VersionList) {
		return confirmedWrites, IncompleteDataError("")
	}

	return confirmedWrites, nil
}

// writeStatusCh - Channel to wait on for peers to write to whether they have confirmed the write
// numRequiredWrites - Needed number of confirmed writes to reach majority
// maxFailures - Number of unconfirmed writes until we should return (otherwise could wait indefinitely for numRequiredWrites)
// Returns a channel with whether the Write was replicated
func CountConfirmedWrites(writeStatusCh chan bool, numRequiredWrites, maxFailures uint8) chan bool {
	writeReplicatedCh := make(chan bool)
	go func() {
		numWrites, numFailures := uint8(0), uint8(0)
		for {
			if numRequiredWrites >= numWrites {
				writeReplicatedCh <- true
			}

			if numFailures > maxFailures {
				writeReplicatedCh <- false
			}
			select {
			case writeSucceeded := <-writeStatusCh:
				if writeSucceeded {
					numWrites++
				} else {
					numFailures++
				}
			}
		}
	}()

	return writeReplicatedCh
}

// Returns error if cannot find data
func GetMissingData(latestVersion int) error {
	if !isSorted {
		sortVersionList()
	}

	// No data has been written
	if latestVersion == 0 {
		return nil
	}

	// Find missing versions
	var missingVersions map[int]bool = make(map[int]bool)
	versionLen := len(VersionList)
	// i is for index in the VersionList
	// m the index we're looking for
	for i, m := FirstMismatch, FirstMismatch; i < latestVersion; i, m = i+1, m+1 {

		// All indices are above our VersionLen so we won't have it
		if i >= versionLen {
			missingVersions[m] = true
			// missingVersions = append(missingVersions, m)
			continue
		}

		for {
			if VersionList[i].Version != m {
				missingVersions[m] = true
				// missingVersions = append(missingVersions, m)
				m++
			} else {
				break
			}
		}
	}

	// Send the list of missingVersions to each Peer
	// The Peer will return a map[versionNum]Data
	// and we will remove the keys that we received from missingVersions
	// We will continue this process until either
	// 1) we have no more missingVersions
	// 2) no more Peers to request data from
	// Case 2 should not happen if there fewer than ClusterSize/2 node failures

	followers := make([]string, 0)
	for ip := range DirectFollowersList {
		followers = append(followers, ip)
	}

	for i := 0; i < len(followers); i++ {
		if len(missingVersions) == 0 {
			break
		}

		ip := followers[i]
		peer, ok := PeerMap.Get(ip)
		if !ok {
			fmt.Println("Peer died")
			continue
		}

		var writeData []FileData
		err := peer.PeerConn.Call("Peer.GetWrites", missingVersions, writeData)
		if err != nil {
			fmt.Println("Err to GetWrites for Peer [%s]", ERR_COL+string(i)+ERR_END)
			continue
		}

		for _, fdata := range writeData {
			delete(missingVersions, fdata.Version)
		}
	}

	if len(missingVersions) != 0 {
		log.Println(ERR_COL + "Data is missing. Could not get data from followers" + ERR_END)
		return IncompleteDataError("")
	}

	return nil
}

///////////////Writing to disk helpers /////////////////
func writeToDisk(path string) error {
	fileData := ClusterData{
		Topic:   TopicName,
		Dataset: VersionList,
	}

	fname := filepath.Join(path, "data.json")
	contents, err := json.MarshalIndent(fileData, "", "  ")
	if err = ioutil.WriteFile(fname, contents, 0644); err != nil {
		log.Println(ERR_COL + "ERROR WRITING TO DISK" + ERR_END)
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
// Returns length of VersionList if all indices match its version number
func sortVersionList() int {
	sort.Slice(VersionList, func(i, j int) bool {
		return VersionList[i].Version > VersionList[j].Version
	})

	var j int
	for i, fdata := range VersionList {
		if i+1 != fdata.Version {
			FirstMismatch = i
			return i
		}
		j++
	}

	FirstMismatch = j + 1
	return j + 1
}

// Returns the longest list of continuous ordered writes
// VersionList: [1,2,3,4,6,7,8] will return [1,2,3,4]
// Note: The caller of this function is responsible for locking
//       since it likely has to do additional work related to the VersionList
func GetConfirmedWrites() []string {
	VersionListLock.Lock()
	defer VersionListLock.Unlock()

	if !isSorted {
		sortVersionList()
	}

	writes := make([]string, 0)
	for _, fdata := range VersionList[:FirstMismatch] {
		writes = append(writes, fdata.Data)
	}
	return writes
}

// Return the highest version number the node has. If node has no data, returns 0
func GetLatestVersion() int {
	VersionListLock.Lock()
	defer VersionListLock.Unlock()

	versionLen := len(VersionList)
	if versionLen == 0 {
		return 0
	}

	if isSorted {
		return VersionList[versionLen-1].Version
	}

	var max int = 0
	for _, fdata := range VersionList[:FirstMismatch] {
		if max < fdata.Version {
			max = fdata.Version
		}
	}

	return max
}

/////////////// End VersionList Helpers ///////////////////
