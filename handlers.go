package main

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/etw/pointapi"
)

type FeedMeta struct {
	Title string
	ID    string
	Href  string
	Self  string
}

type Filter struct {
	Users []string
	Tags  []string
}

type Job struct {
	Rid       string
	Meta      FeedMeta
	Data      []pointapi.PostMeta
	Before    int
	MinPosts  int
	Blacklist *Filter
}

func resRender(res *http.ResponseWriter, job *Job) {
	if feed, err := makeFeed(job); err != nil {
		log.Printf("[ERROR] {%s} Failed to parse point response: %s\n", job.Rid, err)
		(*res).WriteHeader(500)
	} else if result, err := xml.Marshal(feed); err != nil {
		log.Printf("[ERROR] {%s} Failed to render XML: %s\n", job.Rid, err)
		(*res).WriteHeader(500)
	} else {
		(*res).Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		(*res).Header().Set("Request-Id", job.Rid)
		(*res).Write(result)
	}
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.Write(readme)
}

func makeJob(p url.Values) (Job, error) {
	var (
		job Job
		bl  Filter
	)

	rid := make([]byte, 8)
	if _, err := rand.Read(rid); err != nil {
		log.Println("[Error] {0000000000000000} Couldn't generate request id: %s", err)
		return job, err
	} else {
		job.Rid = fmt.Sprintf("%x", rid)
	}

	if val, ok := p["before"]; ok {
		var err error
		if job.Before, err = strconv.Atoi(val[0]); err != nil {
			log.Printf("[WARN] {%s} Couldn't parse 'before' param: %s\n", job.Rid, err)
			return job, err
		}
	} else {
		job.Before = 0
	}

	if val, ok := p["minposts"]; ok {
		var err error
		if job.MinPosts, err = strconv.Atoi(val[0]); err != nil {
			log.Printf("[WARN] {%s} Couldn't parse 'minposts' param: %s\n", job.Rid, err)
			return job, err
		}
	} else {
		job.MinPosts = 20
	}

	if val, ok := p["nouser"]; ok {
		bl.Users = val
		job.Blacklist = &bl
	}
	if val, ok := p["notag"]; ok {
		bl.Tags = val
		job.Blacklist = &bl
	}

	return job, nil
}

func allHandler(res http.ResponseWriter, req *http.Request) {
	var (
		params = req.URL.Query()
		job    Job
		err    error
	)

	if job, err = makeJob(params); err != nil {
		res.WriteHeader(500)
		return
	}
	log.Printf("[INFO] {%s} %s %s\n", job.Rid, req.Method, req.RequestURI)

	job.Meta = FeedMeta{
		Title: "All posts",
		ID:    "https://point.im/all",
		Href:  "https://point.im/all",
		Self:  fmt.Sprintf("http://%s%s", req.Host, req.URL.Path),
	}

	start, has_next := job.Before, true
	for has_next && len(job.Data) < job.MinPosts {
		var data *pointapi.PostList

		log.Printf("[DEBUG] {%s} Requesting posts before: %d\n", job.Rid, start)
		if data, err = api.Point.GetAll(start); err != nil {
			log.Printf("[ERROR] {%s} Failed to get all posts: %s\n", job.Rid, err)
			res.WriteHeader(500)
			return
		}

		start = data.Posts[len(data.Posts)-1].Uid
		has_next = data.HasNext

		job.Data = append(job.Data, filterPosts(data.Posts, job.Blacklist)...)
		log.Printf("[DEBUG] {%s} We have %d posts, need at least %d\n", job.Rid, len(job.Data), job.MinPosts)
	}

	resRender(&res, &job)
}

func tagsHandler(res http.ResponseWriter, req *http.Request) {
	var (
		params = req.URL.Query()
		job    Job
		err    error
	)

	if job, err = makeJob(params); err != nil {
		res.WriteHeader(500)
		return
	}
	log.Printf("[INFO] {%s} %s %s\n", job.Rid, req.Method, req.RequestURI)

	tags := params["tag"]
	sort.Strings(tags)

	if len(tags) < 1 {
		log.Printf("[WARN] {%s} At least one tag is needed\n", job.Rid)
		res.WriteHeader(400)
		fmt.Fprintln(res, "At least one tag is needed")
		return
	}

	job.Meta = FeedMeta{
		Title: fmt.Sprintf("Tagged posts (%s)", strings.Join(tags, ", ")),
		ID:    fmt.Sprintf("https://point.im/?tag=%s", strings.Join(tags, "&tag=")),
		Href:  fmt.Sprintf("https://point.im/?tag=%s", strings.Join(tags, "&tag=")),
		Self:  fmt.Sprintf("http://%s%s?tag=%s", req.Host, req.URL.Path, strings.Join(tags, "&tag=")),
	}

	start, has_next := job.Before, true
	for has_next && len(job.Data) < job.MinPosts {
		var data *pointapi.PostList

		log.Printf("[DEBUG] {%s} Requesting posts before: %d\n", job.Rid, start)
		if data, err = api.Point.GetTags(start, tags); err != nil {
			log.Printf("[ERROR] {%s} Failed to get tagged posts: %s\n", job.Rid, err)
			res.WriteHeader(500)
			return
		}

		start = data.Posts[len(data.Posts)-1].Uid
		has_next = data.HasNext

		job.Data = append(job.Data, filterPosts(data.Posts, job.Blacklist)...)
		log.Printf("[DEBUG] {%s} We have %d posts, need at least %d\n", job.Rid, len(job.Data), job.MinPosts)
	}

	resRender(&res, &job)
}
