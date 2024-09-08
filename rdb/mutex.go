package rdb

import (
	"sync"

	"github.com/go-redsync/redsync"
)

var (
	mutex   sync.RWMutex
	mapSync = map[string]*redsync.Redsync{}
)

func Mutex(dbName, key string) IMutexer {
	return redSync(dbName).NewMutex(key)
}

//  --------------------
func redSync(dbName string) *redsync.Redsync {
	mutex.RLock()
	s, ok := mapSync[dbName]
	mutex.RUnlock()
	if !ok {
		mutex.Lock()
		s = redsync.New([]redsync.Pool{Pool(dbName)})
		mapSync[dbName] = s
		mutex.Unlock()
	}
	return s
}
