package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/etw/pointapi"
)

type FeedMeta struct {
	Title string
	ID    string
	Href  string
}

type Job struct {
	Rid  *string
	Meta *FeedMeta
	Data *pointapi.PostList
	API  *APISet
}

func resRender(res *http.ResponseWriter, job *Job) {
	result, err := makeFeed(job)
	if err != nil {
		log.Printf("[ERROR] Failed to parse point response: %s\n", err)
		(*res).WriteHeader(500)
		return
	}
	feed, err := xml.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to render XML: %s\n", err)
		(*res).WriteHeader(500)
		return
	}
	(*res).Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	(*res).Header().Set("Request-Id", *job.Rid)
	fmt.Fprintln(*res, string(feed))
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "https://github.com/etw/pointfeed/blob/develop/README.md")
}

func allHandler(api *APISet) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var before int
		params := req.URL.Query()

		rid, err := getRid()
		if err != nil {
			res.WriteHeader(500)
			return
		}
		log.Printf("[INFO] {%s} %s %s\n", *rid, req.Method, req.RequestURI)

		b, ok := params["before"]
		if ok {
			before, err = strconv.Atoi(b[0])
			if err != nil {
				log.Printf("[WARN] {%s} Couldn't parse 'before' param: %s\n", *rid, err)
				res.WriteHeader(400)
				return
			}
		} else {
			before = 0
		}

		body, err := api.Point.GetAll(before)
		if err != nil {
			log.Printf("[ERROR] {%s} Failed to get all posts: %s\n", *rid, err)
			res.WriteHeader(500)
			return
		}

		feed := FeedMeta{
			Title: "All posts",
			ID:    "all",
			Href:  "https://point.im/all",
		}

		job := Job{
			Rid:  rid,
			Meta: &feed,
			Data: body,
			API:  api,
		}

		resRender(&res, &job)
	}
}

func tagsHandler(api *APISet) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var before int
		params := req.URL.Query()
		tags := params["tag"]

		rid, err := getRid()
		if err != nil {
			res.WriteHeader(500)
			return
		}
		log.Printf("[INFO] {%s} %s %s\n", *rid, req.Method, req.RequestURI)

		b, ok := params["before"]
		if ok {
			before, err = strconv.Atoi(b[0])
			if err != nil {
				log.Printf("[WARN] {%s} Failed parse before param: %s\n", *rid, err)
				res.WriteHeader(500)
				return
			}
		} else {
			before = 0
		}

		if len(tags) < 1 {
			log.Printf("[WARN] {%s} At least one tag is needed\n", *rid)
			res.WriteHeader(400)
			fmt.Fprintln(res, "At least one tag is needed")
			return
		}

		body, err := api.Point.GetTags(before, tags)
		if err != nil {
			log.Printf("[ERROR] {%s} Failed to get tagged posts: %s\n", *rid, err)
			res.WriteHeader(500)
			return
		}

		sort.Strings(tags)
		feed := FeedMeta{
			Title: fmt.Sprintf("Tagged posts (%s)", strings.Join(tags, ", ")),
			ID:    fmt.Sprintf("tags:%s", strings.Join(tags, ",")),
			Href:  fmt.Sprintf("https://point.im/?tag=%s", strings.Join(tags, "&tag=")),
		}

		job := Job{
			Rid:  rid,
			Meta: &feed,
			Data: body,
			API:  api,
		}

		resRender(&res, &job)
	}
}
