package node

import ("sync"

)

var fm sync.Map

func MountFileMap() error {
	fm = new(sync.Map)
}

func WriteToFileMap(data string) error {
	fm.Store
}