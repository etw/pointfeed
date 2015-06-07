package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strings"

	point "github.com/etw/pointapi"
)

func resRender(res *http.ResponseWriter, job *Job) {
	feed := makeFeed(job)
	if result, err := xml.Marshal(feed); err != nil {
		logger(ERROR, fmt.Sprintf("{%s} Failed to render XML: %s", job.Rid, err))
		(*res).WriteHeader(500)
	} else {
		(*res).Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		(*res).Header().Set("Request-Id", job.Rid)
		(*res).Write(result)
		logger(DEBUG, fmt.Sprintf("{%s} Response sent", job.Rid))
	}
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.Write(readme)
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

	for num, start, has_next := 0, 0, true; has_next && num < job.MinPosts; {
		var data *point.PostList

		logger(DEBUG, fmt.Sprintf("{%s} Requesting posts before: %d", job.Rid, start))
		if data, err = apiset.Point.GetAll(start); err != nil {
			logger(ERROR, fmt.Sprintf("{%s} Failed to get all posts: %s", job.Rid, err))
			res.WriteHeader(500)
			return
		}

		for i, _ := range data.Posts {
			if filterPost(&data.Posts[i], job.Blacklist) {
				job.Group.Add(1)
				num++
				go makeEntry(&data.Posts[i], &job)
			}
		}

		start, has_next = data.Posts[len(data.Posts)-1].Uid, data.HasNext
		logger(DEBUG, fmt.Sprintf("{%s} We have %d posts, need at least %d", job.Rid, num, job.MinPosts))
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

	for num, start, has_next := 0, 0, true; has_next && num < job.MinPosts; {
		var data *point.PostList

		logger(DEBUG, fmt.Sprintf("{%s} Requesting posts before: %d", job.Rid, start))
		if data, err = apiset.Point.GetTags(start, tags); err != nil {
			logger(ERROR, fmt.Sprintf("{%s} Failed to get tagged posts: %s", job.Rid, err))
			res.WriteHeader(500)
			return
		}

		for i, _ := range data.Posts {
			if filterPost(&data.Posts[i], job.Blacklist) {
				job.Group.Add(1)
				num++
				go makeEntry(&data.Posts[i], &job)
			}
		}

		start, has_next = data.Posts[len(data.Posts)-1].Uid, data.HasNext
		logger(DEBUG, fmt.Sprintf("{%s} We have %d posts, need at least %d", job.Rid, num, job.MinPosts))
	}

	resRender(&res, &job)
}
