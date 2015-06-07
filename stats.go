package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const listHref = `<p><a href="%s" rel="noreferrer">%s</a></p>`

type Stats struct {
	Cache StatsCache
}

type StatsCache struct {
	Posts StatsPCache `json:"posts"`
}

type StatsPCache struct {
	Size   int `json:"size"`
	Total  int `json:"total"`
	Hit    int `json:"hit"`
	Missed int `json:"missed"`
}

func statsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(res, listHref, "/stats/cache", "Статистика кеша")
}

func cacheHandler(s *Stats) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(res).Encode(s.Cache)
	}
}
