package file

import "sync"

var mu *sync.RWMutex

func init() {
	mu = &sync.RWMutex{}
}
