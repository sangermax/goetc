package util

import (
	"sync"
)

type BeeMap struct {
	Lock *sync.RWMutex
	BM   map[string]interface{}
}

func NewBeeMap() *BeeMap {
	return &BeeMap{
		Lock: new(sync.RWMutex),
		BM:   make(map[string]interface{}),
	}
}

//Get from maps return the k's value
func (m *BeeMap) Get(k string) interface{} {
	m.Lock.RLock()
	defer m.Lock.RUnlock()
	if val, ok := m.BM[k]; ok {
		return val
	}
	return nil
}

// Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *BeeMap) Set(k string, v interface{}) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	if val, ok := m.BM[k]; !ok {
		m.BM[k] = v
	} else if val != v {
		m.BM[k] = v
	} else {
		return false
	}
	return true
}

func (m *BeeMap) ReSet(k string, v interface{}) bool {
	m.Delete(k)
	m.Set(k, v)
	return true
}

// Returns true if k is exist in the map.
func (m *BeeMap) Check(k string) bool {
	m.Lock.RLock()
	defer m.Lock.RUnlock()
	if _, ok := m.BM[k]; !ok {
		return false
	}
	return true
}

func (m *BeeMap) Delete(k string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	delete(m.BM, k)
}

//获取元素个数
func (m *BeeMap) Size() int {
	m.Lock.RLock()
	defer m.Lock.RUnlock()

	return len(m.BM)
}

//只读第一个
func (m *BeeMap) GetFirst() interface{} {
	m.Lock.RLock()
	defer m.Lock.RUnlock()

	for _, v := range m.BM {
		if v != nil {
			return v
		}
	}

	return nil
}

//返回第一个，且从map中删除
func (m *BeeMap) DetachFirst() (string, interface{}) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for k, v := range m.BM {
		if v != nil {
			delete(m.BM, k)
			return k, v
		}
	}

	return "", nil
}
