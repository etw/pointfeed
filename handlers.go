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
	result, err := makeFeed(job)
	if err != nil {
		log.Printf("[ERROR] {%s} Failed to parse point response: %s\n", job.Rid, err)
		(*res).WriteHeader(500)
		return
	}
	feed, err := xml.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] {%s} Failed to render XML: %s\n", job.Rid, err)
		(*res).WriteHeader(500)
		return
	}
	(*res).Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	(*res).Header().Set("Request-Id", job.Rid)
	fmt.Fprintln(*res, string(feed))
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.Write(rmf)
}

func makeJob(p url.Values) (Job, error) {
	var (
		job Job
		bl  Filter
		err error

		rid     = make([]byte, 8)
		have_bl = false
	)

	_, err = rand.Read(rid)
	if err != nil {
		log.Println("[Error] {%s} Couldn't generate request id: %s", err)
		return job, err
	}
	job.Rid = fmt.Sprintf("%x", rid)

	b, ok := p["before"]
	if ok {
		job.Before, err = strconv.Atoi(b[0])
		if err != nil {
			log.Printf("[WARN] {%s} Couldn't parse 'before' param: %s\n", job.Rid, err)
			return job, err
		}
	} else {
		job.Before = 0
	}

	m, ok := p["minposts"]
	if ok {
		job.MinPosts, err = strconv.Atoi(m[0])
		if err != nil {
			log.Printf("[WARN] {%s} Couldn't parse 'minposts' param: %s\n", job.Rid, err)
			return job, err
		}
	} else {
		job.MinPosts = 20
	}

	nu, ok := p["nouser"]
	if ok {
		bl.Users = nu
		have_bl = true
	}

	nt, ok := p["notag"]
	if ok {
		bl.Tags = nt
		have_bl = true
	}

	if have_bl {
		job.Blacklist = &bl
	} else {
		job.Blacklist = nil
	}

	return job, nil
}

func allHandler(res http.ResponseWriter, req *http.Request) {
	var (
		params   = req.URL.Query()
		has_next = true
	)

	job, err := makeJob(params)
	if err != nil {
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

	for has_next && len(job.Data) < job.MinPosts {
		data, err := api.Point.GetAll(job.Before)
		if err != nil {
			log.Printf("[ERROR] {%s} Failed to get all posts: %s\n", job.Rid, err)
			res.WriteHeader(500)
			return
		}

		has_next = data.HasNext
		job.Data = append(job.Data, filterPosts(data.Posts, job.Blacklist)...)
		log.Printf("[DEBUG] {%s} We have %d posts, need at least %d\n", job.Rid, len(job.Data), job.MinPosts)
	}

	resRender(&res, &job)
}

func tagsHandler(res http.ResponseWriter, req *http.Request) {
	var (
		params   = req.URL.Query()
		has_next = true
	)

	job, err := makeJob(params)
	if err != nil {
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

	for has_next && len(job.Data) < job.MinPosts {
		data, err := api.Point.GetTags(job.Before, tags)
		if err != nil {
			log.Printf("[ERROR] {%s} Failed to get tagged posts: %s\n", job.Rid, err)
			res.WriteHeader(500)
			return
		}

		has_next = data.HasNext
		job.Data = append(job.Data, filterPosts(data.Posts, job.Blacklist)...)
		log.Printf("[DEBUG] {%s} We have %d posts, need at least %d\n", job.Rid, len(job.Data), job.MinPosts)
	}

	resRender(&res, &job)

}
