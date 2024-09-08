package util

import "sync"

// 普通锁
type ILocker interface {

	// 主锁
	Lock() interface{}

	UnLock(i interface{})
}

// 读写锁
type IRWLocker interface {

	// 写锁
	ILocker

	// 读锁
	RLock() interface{}

	RUnLock(i interface{})
}

//  ==================== mutex lock
func AutoLock(mu *sync.Mutex, fun func()) {
	mu.Lock()
	fun()
	mu.Unlock()
}
func AutoRLock(mu *sync.RWMutex, fun func()) {
	mu.RLock()
	fun()
	mu.RUnlock()
}
func AutoWLock(mu *sync.RWMutex, fun func()) {
	mu.Lock()
	fun()
	mu.Unlock()
}
func Lock(mu *sync.Mutex) *sync.Mutex { mu.Lock(); return mu }

func UnLock(lockFun *sync.Mutex) { lockFun.Unlock() }

func RLock(mu *sync.RWMutex) *sync.RWMutex { mu.RLock(); return mu }

func RUnLock(rMutex *sync.RWMutex) { rMutex.RUnlock() }

func WLock(mu *sync.RWMutex) *sync.RWMutex { mu.Lock(); return mu }

func WUnLock(wMutex *sync.RWMutex) { wMutex.Unlock() }

// ==================== Locker
type Locker struct {
	mutex sync.Mutex
}

func (l *Locker) Mutex() *sync.Mutex { return &l.mutex }

func (l *Locker) Lock() interface{} { l.mutex.Lock(); return nil }

func (l *Locker) UnLock(i interface{}) { l.mutex.Unlock() }

type RWLocker struct {
	mutex sync.RWMutex
}

func (l *RWLocker) Mutex() *sync.RWMutex { return &l.mutex }

func (l *RWLocker) Lock() interface{} { l.mutex.Lock(); return nil }

func (l *RWLocker) UnLock(i interface{}) { l.mutex.Unlock() }

func (l *RWLocker) RLock() interface{} { l.mutex.RLock(); return nil }

func (l *RWLocker) RUnLock(i interface{}) { l.mutex.RUnlock() }

// InvalidLocker is a fake locker
type InvalidLocker struct {
}

// Lock does nothing
func (l InvalidLocker) Lock() interface{} {
	return nil
}

// Unlock does nothing
func (l InvalidLocker) UnLock(i interface{}) {

}

// RLock does nothing
func (l InvalidLocker) RLock() interface{} {
	return nil
}

// RUnlock does nothing
func (l InvalidLocker) RUnLock(i interface{}) {

}
