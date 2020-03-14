package cache

import (
	"kubesphere.io/kubesphere/pkg/server/errors"
	"regexp"
	"strings"
	"time"
)

var ErrNoSuchKey = errors.New("no such key")

type simpleObject struct {
	value  string
	neverExpire bool
	expiredAt time.Time
}

// SimpleCache implements cache.Interface use memory objects, it should be used only for testing
type simpleCache struct {
	store map[string]simpleObject
}

func NewSimpleCache() Interface {
	return &simpleCache{store: make(map[string]simpleObject)}
}

func (s *simpleCache) Keys(pattern string) ([]string, error) {
	// There is a little difference between go regexp and redis key pattern
	// In redis, * means any character, while in go . means match everything.
	pattern = strings.Replace(pattern, "*", ".", -1)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var keys []string
	for k, _ := range s.store {
		if re.MatchString(k) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

func (s *simpleCache) Set(key string, value string, duration time.Duration) error {
	sobject := simpleObject{
		value: value,
		neverExpire: false,
		expiredAt: time.Now().Add(duration),
	}

	if duration == NeverExpire {
		sobject.neverExpire = true
	}

	s.store[key] = sobject
	return nil
}

func (s *simpleCache) Del(keys ...string) error {
	for _, key := range keys {
		if _, ok := s.store[key]; ok {
			delete(s.store, key)
		} else {
			return ErrNoSuchKey
		}
	}

	return nil
}

func (s *simpleCache) Get(key string) (string, error) {
	if sobject, ok := s.store[key]; ok {
		if sobject.neverExpire || time.Now().Before(sobject.expiredAt) {
			return sobject.value, nil
		}
	}

	return "", ErrNoSuchKey
}

func (s *simpleCache) Exists(keys ...string) (bool, error) {
	for _, key := range keys {
		if _, ok := s.store[key]; !ok {
			return false, ErrNoSuchKey
		}
	}

	return true, nil
}

func (s *simpleCache) Expire(key string, duration time.Duration) error {
	value, err := s.Get(key)
	if err != nil {
		return err
	}

	sobject := simpleObject{
		value:  value,
		neverExpire: false,
		expiredAt: time.Now().Add(duration),
	}

	if duration == NeverExpire {
		sobject.neverExpire = true
	}

	s.store[key] = sobject
	return nil
}
