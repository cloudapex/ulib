package util

import (
	"errors"
	"reflect"
)

var (
	defaultLocker InvalidLocker
)

var ErrorNotFound = errors.New("not found")

// Options holds Map's options
type Options struct {
	locker IRWLocker
}

// Option is a function type used to set Options
type Option func(option *Options)

// HashMapWithSafe is used to set a map goroutine-safe
// Note that iterators are not goroutine safe, and it is useless to turn on the setting option here.
// so don't use iterator in multi goroutines
func HashMapWithSafe() Option {
	return func(option *Options) {
		option.locker = &RWLocker{}
	}
}

// HashMap uses RbTress for internal data structure, and every key can must bee unique.
type HashMap[K comparable, V any] struct {
	data   map[K]V
	locker IRWLocker
}

// New creates a new map
func NewHashMap[K comparable, V any](opts ...Option) *HashMap[K, V] {
	option := Options{
		locker: defaultLocker,
	}
	for _, opt := range opts {
		opt(&option)
	}
	return &HashMap[K, V]{
		data:   map[K]V{},
		locker: option.locker,
	}
}

// Insert inserts a key-value to the map
func (m *HashMap[K, V]) Insert(key K, value V) {
	defer m.locker.UnLock(m.locker.Lock())

	m.data[key] = value
}
func (m *HashMap[K, V]) Inserts(pairs ...*Pair[K, V]) {
	defer m.locker.UnLock(m.locker.Lock())

	for _, it := range pairs {
		m.data[it.first] = it.second
	}
}
func (m *HashMap[K, V]) Replace(data map[K]V) {
	defer m.locker.UnLock(m.locker.Lock())
	m.data = data
}

// Get returns the value of the passed key if the key is in the map, otherwise returns nil
func (m *HashMap[K, V]) Get(key K) (V, bool) {
	defer m.locker.RUnLock(m.locker.RLock())

	val, ok := m.data[key]
	return val, ok
}

// Got returns the value of the passed key if the key is in the map, otherwise returns nil
func (m *HashMap[K, V]) Got(key K) V {
	defer m.locker.RUnLock(m.locker.RLock())

	return m.data[key]
}

// Find finds a node by the passed key and returns its iterator
func (m *HashMap[K, V]) Find(key K) *Pair[K, V] {
	defer m.locker.RUnLock(m.locker.RLock())

	value, ok := m.data[key]
	if !ok {
		return nil
	}
	return &Pair[K, V]{key, value}
}

// Keys return []K
func (m *HashMap[K, V]) Keys() []K {
	defer m.locker.RUnLock(m.locker.RLock())
	return Keys(m.data)
}

// Values return []V
func (m *HashMap[K, V]) Values() []V {
	defer m.locker.RUnLock(m.locker.RLock())
	return Values(m.data)
}

// Erase erases the node by the passed key from the map if the key in the Map
func (m *HashMap[K, V]) Erase(key K) {
	defer m.locker.UnLock(m.locker.Lock())

	delete(m.data, key)
}

// Clear clears the map
func (m *HashMap[K, V]) Clear() {
	defer m.locker.UnLock(m.locker.Lock())

	m.data = map[K]V{}
}

// Contains returns true if the key is in the map. otherwise returns false.
func (m *HashMap[K, V]) Contains(keys ...K) bool {
	defer m.locker.RUnLock(m.locker.RLock())

	for _, key := range keys {
		if _, ok := m.data[key]; !ok {
			return false
		}
	}
	return true
}

// Size returns the amount of elements in the map
func (m *HashMap[K, V]) Size() int {
	defer m.locker.RUnLock(m.locker.RLock())

	return len(m.data)
}

func (m *HashMap[K, V]) Iterator(iterFunc func(key K, val V)) {
	defer m.locker.RUnLock(m.locker.RLock())

	for k, v := range m.data {
		iterFunc(k, v)
	}
}

// ------------------------------------------------------

type Pair[K comparable, V any] struct {
	first  K
	second V
}

func MakePair[K comparable, V any](key K, val V) *Pair[K, V] {
	return &Pair[K, V]{first: key, second: val}
}

func (p *Pair[K, V]) Set(key K, val V) {
	p.first, p.second = key, val
}
func (p *Pair[K, V]) Equal(other *Pair[K, V]) bool {
	return p.first == other.first && reflect.DeepEqual(p.second, other.second)
}

func (p *Pair[K, V]) First() K {
	return p.first
}

func (p *Pair[K, V]) Second() V {
	return p.second
}

// -------------------------------------------------------

// Keys returns the keys of the map m.
// The keys will be in an indeterminate order.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

// Values returns the values of the map m.
// The values will be in an indeterminate order.
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

// Equal reports whether two maps contain the same key/value pairs.
// Values are compared using ==.
func Equal[M1, M2 ~map[K]V, K, V comparable](m1 M1, m2 M2) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || v1 != v2 {
			return false
		}
	}
	return true
}

// EqualFunc is like Equal, but compares values using eq.
// Keys are still compared with ==.
func EqualFunc[M1 ~map[K]V1, M2 ~map[K]V2, K comparable, V1, V2 any](m1 M1, m2 M2, eq func(V1, V2) bool) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !eq(v1, v2) {
			return false
		}
	}
	return true
}

// Clear removes all entries from m, leaving it empty.
func Clear[M ~map[K]V, K comparable, V any](m M) {
	for k := range m {
		delete(m, k)
	}
}

// Clone returns a copy of m.  This is a shallow clone:
// the new keys and values are set using ordinary assignment.
func Clone[M ~map[K]V, K comparable, V any](m M) M {
	// Preserve nil in case it matters.
	if m == nil {
		return nil
	}
	r := make(M, len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

// Copy copies all key/value pairs in src adding them to dst.
// When a key in src is already present in dst,
// the value in dst will be overwritten by the value associated
// with the key in src.
func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, v := range src {
		dst[k] = v
	}
}

// DeleteFunc deletes any key/value pairs from m for which del returns true.
func DeleteFunc[M ~map[K]V, K comparable, V any](m M, del func(K, V) bool) {
	for k, v := range m {
		if del(k, v) {
			delete(m, k)
		}
	}
}
