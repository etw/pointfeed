package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/etw/pointapi"
	"golang.org/x/tools/blog/atom"
)

const maxtitle = 96

func makeEntry(e *pointapi.PostMeta, api *APISet) (*atom.Entry, error) {
	var title string

	log.Printf("[DEBUG] Got post; id: %s, author: %s, files: %d\n", e.Post.Id, e.Post.Author.Login, len(e.Post.Files))

	person := atom.Person{
		Name: e.Post.Author.Login,
		URI:  fmt.Sprintf("https://%s.point.im/", e.Post.Author.Login),
	}

	htmlPost, err := renderPost(&e.Post, api)
	if err != nil {
		return nil, errors.New("Couldn't render post in HTML")
	}

	post := atom.Text{
		Type: "html",
		Body: htmlPost,
	}

	runestr := []rune(e.Post.Text)
	nl := findNl(runestr)

	if nl > maxtitle || (nl < 0 && len(runestr) > maxtitle) {
		title = fmt.Sprintf("%s...", string(runestr[:(maxtitle-3)]))
	} else {
		title = string(runestr)
	}

	entry := atom.Entry{
		Title: title,
		ID:    fmt.Sprintf("%d", e.Uid),
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
	//f *FeedMeta, p []pointapi.PostMeta
	var posts []*atom.Entry
	var timestamp atom.TimeStr

	for i := range job.Data.Posts {
		entry, err := makeEntry(&job.Data.Posts[i], job.API)
		if err != nil {
			return nil, err
		}
		posts = append(posts, entry)
	}

	if len(job.Data.Posts) > 0 {
		timestamp = atom.Time(job.Data.Posts[0].Post.Created)
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
		},
		Updated: timestamp,
		Entry:   posts,
	}

	return &feed, nil
}
