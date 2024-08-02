package cache

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Cache interface {
	Put(key string, value interface{}, expire time.Duration) error

	Delete(key string) (interface{}, error)

	Get(key string) (interface{}, error)
}

func NewCache() Cache {
	return &MapCache{
		cache: map[string]interface{}{},
	}
}
func NewValueCache() *IDValueCache {
	return &IDValueCache{
		cache: NewCache(),
	}
}

type IDValueCache struct {
	cache Cache
}

func getKey(value interface{}) (string, error) {

	v := reflect.ValueOf(value)
	v = v.Elem()
	t := v.Type()

	valueName := t.Name()
	idField := v.FieldByName("ID")
	if !idField.IsValid() || idField.Kind() != reflect.Uint {
		return "", errors.New("model must have an ID field of an integer type")
	}

	id := idField.Uint()
	Key := fmt.Sprintf("%s:%d", valueName, id)
	return Key, nil

}

func (pc *IDValueCache) Put(value interface{}, expire time.Duration) error {

	key, err := getKey(value)
	if err != nil {
		return err
	}

	return pc.cache.Put(key, value, expire)
}

func (pc *IDValueCache) Delete(value interface{}) (interface{}, error) {
	key, err := getKey(value)
	if err != nil {
		return nil, err
	}

	return pc.cache.Delete(key)
}

func (pc *IDValueCache) Get(value interface{}) (interface{}, error) {
	key, err := getKey(value)
	if err != nil {
		return nil, err
	}

	return pc.cache.Get(key)
}
