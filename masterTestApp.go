/*
File masterTestApp.go is the master test app. It starts all of the other test
apps and receives the co-ordinates that they generate directly. It uses this to
generate a heatmap.

A different app will be used to receive data from the distributed data queues.
*/
package main

import (
	"./lib/producer"
	"./movement"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var collectedPoints []movement.Point
var collectedPointsLock sync.Mutex

func main() {
	fmt.Println("This program connects to the data service as well as to an internal port for")
	fmt.Println("sending data to a webserver.")
	fmt.Println("Usage: go run <file> <serv-ip>:<serv-port> <internal-ip:internal-port>")
	var files []string

	root := "./testGraphs"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	internalConn, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		fmt.Println("Could not connect to internal:", err)
	}

	var producerNodeId = 0
	for i, file := range files {
		if i == 0 {
			continue
		}

		fmt.Println("Starting client node with graph file", file)

		m, err, firstPoint := movement.CreateLocationGraph(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		myId := producerNodeId
		wSess, err := producer.OpenTopic("gps_coords", os.Args[1], fmt.Sprintf("Writer %d", myId))
		if err != nil {
			continue
		}

		fn := func(p movement.Point) {
			fmt.Printf("ID%d: %.2f %.2f\n", myId, p.X, p.Y)
			appendPoint(p)

			datum := fmt.Sprintf("%.2f %.2f\n", p.X, p.Y)

			internalConn.Write([]byte(datum))
			wSess.Write(datum)
		}

		producerNodeId++

		// parse the speed from the filename
		f, err := strconv.ParseFloat(file[strings.Index(file, "/")+1:], 64)
		if err != nil {
			continue
		}
		go movement.Travel(firstPoint, m, f, fn)
	}

	time.Sleep(10 * time.Second)
}

// Use this to atomically append a point to the global point slice
func appendPoint(point movement.Point) {
	collectedPointsLock.Lock()
	defer collectedPointsLock.Unlock()

	collectedPoints = append(collectedPoints, point)
}

// Self explanatory
func printAllPoints() {
	collectedPointsLock.Lock()
	defer collectedPointsLock.Unlock()

	for _, p := range collectedPoints {
		fmt.Printf("%.2f %.2f\n", p.X, p.Y)
	}
}
