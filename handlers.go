package main

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	point "github.com/etw/pointapi"
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
	Data      []point.PostMeta
	MinPosts  int
	Blacklist *Filter
}

func resRender(res *http.ResponseWriter, job *Job) {
	if feed, err := makeFeed(job); err != nil {
		logger(ERROR, fmt.Sprintf("{%s} Failed to parse point response: %s", job.Rid, err))
		(*res).WriteHeader(500)
	} else if result, err := xml.Marshal(feed); err != nil {
		logger(ERROR, fmt.Sprintf("{%s} Failed to render XML: %s", job.Rid, err))
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

func cacheLister(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "/cache/posts\n")
}

func cacheHandler(s *Stats) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		s.render(&res)
	}
}

func makeJob(p url.Values) (Job, error) {
	var (
		job Job
		bl  Filter
	)

	rid := make([]byte, 8)
	if _, err := rand.Read(rid); err != nil {
		logger(ERROR, fmt.Sprintf("{0000000000000000} Couldn't generate request id: %s", err))
		return job, err
	} else {
		job.Rid = fmt.Sprintf("%x", rid)
	}

	if val, ok := p["minposts"]; ok {
		var err error
		if job.MinPosts, err = strconv.Atoi(val[0]); err != nil {
			logger(WARN, fmt.Sprintf("{%s} Couldn't parse 'minposts' param: %s", job.Rid, err))
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
	logger(INFO, fmt.Sprintf("{%s} %s %s", job.Rid, req.Method, req.RequestURI))

	job.Meta = FeedMeta{
		Title: "All posts",
		ID:    "https://point.im/all",
		Href:  "https://point.im/all",
		Self:  fmt.Sprintf("http://%s%s", req.Host, req.URL.Path),
	}

	start, has_next := 0, true
	for has_next && len(job.Data) < job.MinPosts {
		var data *point.PostList

		logger(DEBUG, fmt.Sprintf("{%s} Requesting posts before: %d", job.Rid, start))
		if data, err = apiset.Point.GetAll(start); err != nil {
			logger(ERROR, fmt.Sprintf("{%s} Failed to get all posts: %s", job.Rid, err))
			res.WriteHeader(500)
			return
		}

		start = data.Posts[len(data.Posts)-1].Uid
		has_next = data.HasNext

		job.Data = append(job.Data, filterPosts(data.Posts, job.Blacklist)...)
		logger(DEBUG, fmt.Sprintf("{%s} We have %d posts, need at least %d", job.Rid, len(job.Data), job.MinPosts))
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
	logger(INFO, fmt.Sprintf("{%s} %s %s", job.Rid, req.Method, req.RequestURI))

	tags := params["tag"]
	sort.Strings(tags)

	if len(tags) < 1 {
		logger(WARN, fmt.Sprintf("{%s} At least one tag is needed", job.Rid))
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

	start, has_next := 0, true
	for has_next && len(job.Data) < job.MinPosts {
		var data *point.PostList

		logger(DEBUG, fmt.Sprintf("{%s} Requesting posts before: %d", job.Rid, start))
		if data, err = apiset.Point.GetTags(start, tags); err != nil {
			logger(ERROR, fmt.Sprintf("{%s} Failed to get tagged posts: %s", job.Rid, err))
			res.WriteHeader(500)
			return
		}

		start = data.Posts[len(data.Posts)-1].Uid
		has_next = data.HasNext

		job.Data = append(job.Data, filterPosts(data.Posts, job.Blacklist)...)
		logger(DEBUG, fmt.Sprintf("{%s} We have %d posts, need at least %d", job.Rid, len(job.Data), job.MinPosts))
	}

	resRender(&res, &job)
}
