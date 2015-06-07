package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

const pCacheSize = 256

type Stats struct {
	Size   int `json:"size"`
	Total  int `json:"total"`
	Hit    int `json:"hit"`
	Missed int `json:"missed"`
}

func (s *Stats) render(w *http.ResponseWriter) {
	json.NewEncoder(*w).Encode(s)
}

func pGet(uid string) func() (interface{}, error) {
	return func() (interface{}, error) {
		pStats.Total++
		if c, ok := pCache.Get(uid); ok {
			pStats.Hit++
			return c, nil
		} else {
			pStats.Missed++
			return nil, errors.New("Cache miss")
		}
	}
}

func pPut(uid string, buf []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		pCache.Add(uid, buf)
		pStats.Size = pCache.Len()
		return nil, nil
	}
}
