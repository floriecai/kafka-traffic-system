package node

import ("sync"
		"../../structs"
		"fmt"
)

var fm *sync.Map


// FileSystem related errors //////
type FileSystemError struct {
	Reason string
}

func (e FileSystemError) Error() string {
	return fmt.Sprintf(e.Reason)
}
////////////////////////////////////

func MountFile() error {
	fm = new(sync.Map)
	// TODO: Make fault tolerant by reading files from drive
	return nil
}

func WriteFile(coords structs.Coords) error {
	fm.Store(coords.Id, coords.Data)
	//TODO: Write to Drive as well
	return nil
}

func ReadFile(req int) (string, error) {
	data, ok := fm.Load(req)
	if ok {
		coords := data.(string)
		return coords, nil
	} else {
		return "", FileSystemError{"a general error"}
	}
}