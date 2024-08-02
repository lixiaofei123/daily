package cache

import "time"

type MapCache struct {
	cache map[string]interface{}
}

func (m *MapCache) Put(key string, value interface{}, expire time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *MapCache) Delete(key string) (interface{}, error) {
	if oldvalue, ok := m.cache[key]; ok {
		delete(m.cache, key)
		return oldvalue, nil
	}
	return nil, nil
}

func (m *MapCache) Get(key string) (interface{}, error) {
	if cachevalue, ok := m.cache[key]; ok {
		return cachevalue, nil
	}
	return nil, nil
}
