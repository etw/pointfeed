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
}

type Job struct {
	Rid     string
	Meta    FeedMeta
	Data    *pointapi.PostList
	Before  int
	MinSize int
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
	fmt.Fprint(res, "https://github.com/etw/pointfeed/blob/develop/README.md")
}

func makeJob(p url.Values) (Job, error) {
	var (
		job Job
		err error
		rid = make([]byte, 8)
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
	return job, nil
}

func allHandler(res http.ResponseWriter, req *http.Request) {
	var params = req.URL.Query()

	job, err := makeJob(params)
	if err != nil {
		res.WriteHeader(500)
		return
	}
	log.Printf("[INFO] {%s} %s %s\n", job.Rid, req.Method, req.RequestURI)

	job.Data, err = api.Point.GetAll(job.Before)
	if err != nil {
		log.Printf("[ERROR] {%s} Failed to get all posts: %s\n", job.Rid, err)
		res.WriteHeader(500)
		return
	}

	job.Meta = FeedMeta{
		Title: "All posts",
		ID:    "all",
		Href:  "https://point.im/all",
	}

	resRender(&res, &job)
}

func tagsHandler(res http.ResponseWriter, req *http.Request) {
	var params = req.URL.Query()

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

	job.Data, err = api.Point.GetTags(job.Before, tags)
	if err != nil {
		log.Printf("[ERROR] {%s} Failed to get tagged posts: %s\n", job.Rid, err)
		res.WriteHeader(500)
		return
	}

	job.Meta = FeedMeta{
		Title: fmt.Sprintf("Tagged posts (%s)", strings.Join(tags, ", ")),
		ID:    fmt.Sprintf("tags:%s", strings.Join(tags, ",")),
		Href:  fmt.Sprintf("https://point.im/?tag=%s", strings.Join(tags, "&tag=")),
	}

	resRender(&res, &job)

}
