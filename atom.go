package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	point "github.com/etw/pointapi"

	"golang.org/x/tools/blog/atom"
)

const maxTitle = 96

func makeEntry(p *point.PostMeta, job *Job) (*atom.Entry, error) {
	var (
		title string
		body = new(bytes.Buffer)
	)

	logger(DEBUG, fmt.Sprintf("{%s} Got post; id: %s, author: %s, files: %d", job.Rid, p.Post.Id, p.Post.Author.Login, len(p.Post.Files)))

	if c, err := doGroup.Do("posts", pGet(p.Post.Id)); err != nil {
		if err := renderPost(body, &p.Post); err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't render post: %s", err))
		} else {
			logger(DEBUG, fmt.Sprintf("{%s} Cache miss (uid %s, csize %d)", job.Rid, p.Post.Id, pCache.Len()))
			doGroup.Do("posts", pPut(p.Post.Id, body.Bytes()))
		}
	} else {
		logger(DEBUG, fmt.Sprintf("{%s} Cache hit (uid %s, csize %d)", job.Rid, p.Post.Id, pCache.Len()))
		body.Write(c.([]byte))
	}

	defer body.Reset()

	person := atom.Person{
		Name: p.Post.Author.Login,
		URI:  fmt.Sprintf("https://%s.point.im/", p.Post.Author.Login),
	}

	post := atom.Text{
		Type: "html",
		Body: body.String(),
	}

	runestr := []rune(p.Post.Text)
	nl := findNl(runestr)

	if nl > maxTitle || (nl < 0 && len(runestr) > maxTitle) {
		title = fmt.Sprintf("%s...", string(runestr[:(maxTitle-3)]))
	} else if nl >= 0 && nl <= maxTitle {
		title = string(runestr[:nl])
	} else {
		title = string(runestr)
	}

	entry := &atom.Entry{
		Title: title,
		ID:    fmt.Sprintf("%s/%s", point.POINTIM, p.Post.Id),
		Link: []atom.Link{
			atom.Link{
				Rel:  "alternate",
				Href: fmt.Sprintf("%s/%s", point.POINTIM, p.Post.Id),
			},
		},
		Published: atom.Time(p.Post.Created),
		Updated:   atom.Time(p.Post.Created),
		Author:    &person,
		Content:   &post,
	}
	return entry, nil
}

func makeFeed(job *Job) (*atom.Feed, error) {
	var posts []*atom.Entry
	var timestamp atom.TimeStr

	for i := range job.Data {
		if entry, err := makeEntry(&job.Data[i], job); err != nil {
			return nil, err
		} else {
			posts = append(posts, entry)
		}
	}

	if len(job.Data) > 0 {
		timestamp = atom.Time(job.Data[0].Post.Created)
	} else {
		timestamp = atom.Time(time.Now())
	}

	feed := atom.Feed{
		Title: job.Meta.Title,
		ID:    job.Meta.ID,
		Link: []atom.Link{
			atom.Link{
				Rel:  "alternate",
				Href: job.Meta.Href,
			},
			atom.Link{
				Rel:  "self",
				Href: job.Meta.Self,
			},
		},
		Updated: timestamp,
		Entry:   posts,
	}

	return &feed, nil
}
