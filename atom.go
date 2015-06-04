package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/etw/pointapi"
	"golang.org/x/tools/blog/atom"
)

const maxTitle = 96

func makeEntry(e *pointapi.PostMeta) (*atom.Entry, error) {
	var title string
	var body bytes.Buffer

	log.Printf("[DEBUG] Got post; id: %s, author: %s, files: %d\n", e.Post.Id, e.Post.Author.Login, len(e.Post.Files))

	person := atom.Person{
		Name: e.Post.Author.Login,
		URI:  fmt.Sprintf("https://%s.point.im/", e.Post.Author.Login),
	}

	if renderPost(&body, &e.Post) != nil {
		return nil, errors.New("Couldn't render post")
	}
	defer body.Reset()

	post := atom.Text{
		Type: "html",
		Body: body.String(),
	}

	runestr := []rune(e.Post.Text)
	nl := findNl(runestr)

	if nl > maxTitle || (nl < 0 && len(runestr) > maxTitle) {
		title = fmt.Sprintf("%s...", string(runestr[:(maxTitle-3)]))
	} else if nl >= 0 && nl <= maxTitle {
		title = string(runestr[:nl])
	} else {
		title = string(runestr)
	}

	entry := atom.Entry{
		Title: title,
		ID:    fmt.Sprintf("https://point.im/%s", e.Post.Id),
		Link: []atom.Link{
			atom.Link{
				Rel:  "alternate",
				Href: fmt.Sprintf("https://point.im/%s", e.Post.Id),
			},
		},
		Published: atom.Time(e.Post.Created),
		Updated:   atom.Time(e.Post.Created),
		Author:    &person,
		Content:   &post,
	}
	return &entry, nil
}

func makeFeed(job *Job) (*atom.Feed, error) {
	var posts []*atom.Entry
	var timestamp atom.TimeStr

	for i := range job.Data {
		if entry, err := makeEntry(&job.Data[i]); err != nil {
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
