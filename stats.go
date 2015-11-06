package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const listHref = `<p><a href="%s" rel="noreferrer">%s</a></p>`

type Stats struct {
	Cache StatsCache
	Media StatsMedia
}

type StatsMedia struct {
	Gelbooru uint64 `json:"gelbooru"`
	Danbooru uint64 `json:"danbooru"`
	Youtube  uint64 `json:"youtube"`
	Coub     uint64 `json:"coub"`
	Image    uint64 `json:"images"`
	Audio    uint64 `json:"audio"`
	Video    uint64 `json:"video"`
	Https    uint64 `json:"https"`
}

type StatsCache struct {
	Posts StatsPCache `json:"posts"`
}

type StatsPCache struct {
	Size   int    `json:"size"`
	Total  uint64 `json:"total"`
	Hit    uint64 `json:"hit"`
	Missed uint64 `json:"missed"`
}

func statsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(res, listHref, "/stats/cache", "Статистика кеша")
	fmt.Fprintf(res, listHref, "/stats/media", "Статистика по обработке URL-ов")
}

func cacheHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(res).Encode(stats.Cache)
}

func mediaHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(res).Encode(stats.Media)
}
