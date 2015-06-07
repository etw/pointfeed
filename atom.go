package main

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	point "github.com/etw/pointapi"

	"golang.org/x/tools/blog/atom"
)

const maxTitle = 96

type Entry struct {
	Atom      *atom.Entry
	Timestamp *time.Time
}

type Entries []*Entry

func (e Entries) Atom() []*atom.Entry {
	var r []*atom.Entry
	for _, v := range e {
		r = append(r, v.Atom)
	}
	return r
}

func (e Entries) Len() int {
	return len(e)
}

func (e Entries) Less(i, j int) bool {
	if e[i].Timestamp.After(*e[j].Timestamp) {
		return true
	}
	return false
}

func (e Entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (job *Job) makeEntry(p *point.PostMeta) {
	var (
		title string
		body  = new(bytes.Buffer)
	)

	logger(DEBUG, fmt.Sprintf("{%s} Got post; id: %s, author: %s, files: %d", job.Rid, p.Post.Id, p.Post.Author.Login, len(p.Post.Files)))

	if c, err := doGroup.Do("posts", pGet(p.Post.Id)); err != nil {
		if err := renderPost(body, &p.Post); err != nil {
			logger(ERROR, fmt.Sprintf("Couldn't render post: %s", err))
			return
		} else {
			doGroup.Do("posts", pPut(p.Post.Id, body.Bytes()))
			logger(DEBUG, fmt.Sprintf("{%s} PCache miss (uid %s, csize %d)", job.Rid, p.Post.Id, pCache.Len()))
		}
	} else {
		logger(DEBUG, fmt.Sprintf("{%s} PCache hit (uid %s, csize %d)", job.Rid, p.Post.Id, pCache.Len()))
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

	job.Queue <- &Entry{
		Atom: &atom.Entry{
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
		},
		Timestamp: &p.Post.Created,
	}
	job.Group.Done()
}

func (job *Job) makeFeed() *atom.Feed {
	var (
		posts     []*Entry
		timestamp atom.TimeStr
	)
	defer close(job.Queue)

	job.Group.Wait()
	for len(job.Queue) > 0 {
		posts = append(posts, <-job.Queue)
	}
	sort.Sort(Entries(posts))

	if len(posts) > 0 {
		timestamp = (*posts[0]).Atom.Published
	} else {
		timestamp = atom.Time(time.Now())
	}

	return &atom.Feed{
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
		Entry:   Entries(posts).Atom(),
	}
}
