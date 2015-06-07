package main

import (
	"errors"
)

const pCacheSize = 8192

func pGet(uid string) func() (interface{}, error) {
	return func() (interface{}, error) {
		stats.Cache.Posts.Total++
		if c, ok := pCache.Get(uid); ok {
			stats.Cache.Posts.Hit++
			return c, nil
		} else {
			stats.Cache.Posts.Missed++
			return nil, errors.New("Cache miss")
		}
	}
}

func pPut(uid string, buf []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		pCache.Add(uid, buf)
		stats.Cache.Posts.Size = pCache.Len()
		return nil, nil
	}
}
