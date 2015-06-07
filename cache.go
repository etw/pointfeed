package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Stats struct {
	Size   int `json:"size"`
	Total  int `json:"total"`
	Hit    int `json:"hit"`
	Missed int `json:"missed"`
}

func (s *Stats) renderStats(w *http.ResponseWriter) {
	json.NewEncoder(*w).Encode(s)
}

func pGet(uid int) func() (interface{}, error) {
	return func() (interface{}, error) {
		if c, ok := pCache.Get(uid); ok {
			pStats.Hit++
			pStats.Total++
			return c, nil
		} else {
			pStats.Missed++
			pStats.Total++
			return nil, errors.New("Cache miss")
		}
	}
}

func pPut(uid int, buf []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		pCache.Add(uid, buf)
		pStats.Size = pCache.Len()
		return nil, nil
	}
}
