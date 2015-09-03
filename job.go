package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	point "github.com/etw/pointapi"
)

const chanSize = 20

type FeedMeta struct {
	Title string
	ID    string
	Href  string
	Self  string
}

type Filter struct {
	NoUsers []string
	NoTags  []string
	AndTags []string
}

type Job struct {
	Rid      string
	Meta     *FeedMeta
	Queue    chan *Entry
	Workers  int
	MinPosts int
	Bnwlist  *Filter
}

func newJob(p url.Values) (Job, error) {
	var (
		job Job
		fil Filter
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
		fil.NoUsers = val
		job.Bnwlist = &fil
	}
	if val, ok := p["notag"]; ok {
		fil.NoTags = val
		job.Bnwlist = &fil
	}
	if val, ok := p["andtag"]; ok {
		fil.AndTags = val
		job.Bnwlist = &fil
	}

	job.Workers = 0
	job.Queue = make(chan *Entry, chanSize)

	return job, nil
}

func (job *Job) procJob(res *http.ResponseWriter, fn func(int) (*point.PostList, error)) {
	for start, has_next := 0, true; has_next && job.Workers < job.MinPosts; {
		logger(DEBUG, fmt.Sprintf("{%s} Requesting posts before: %d", job.Rid, start))
		if data, err := fn(start); err != nil {
			(*res).WriteHeader(500)
			return
		} else {
			for i, _ := range data.Posts {
				if filterPost(&data.Posts[i], job.Bnwlist) {
					job.Workers++
					go job.makeEntry(&data.Posts[i])
				}
			}
			start, has_next = data.Posts[len(data.Posts)-1].Uid, data.HasNext
			logger(DEBUG, fmt.Sprintf("{%s} We have %d posts, need at least %d", job.Rid, job.Workers, job.MinPosts))
		}
	}
	job.resRender(res)
}
