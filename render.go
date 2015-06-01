package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"golang.org/x/tools/blog/atom"

	"github.com/etw/pointapi"
)

const maxtitle = 96

func renderPost(p *string) (*string, error) {
	r := strings.Replace(*p, "\n", "<br>", -1)
	return &r, nil
}

func makeEntry(e *pointapi.PostMeta) (*atom.Entry, error) {
	var title string

	person := atom.Person{
		Name: e.Post.Author.Login,
		URI:  fmt.Sprintf("https://%s.point.im/", e.Post.Author.Login),
	}

	escPost := html.EscapeString(e.Post.Text)
	htmlPost, err := renderPost(&escPost)
	if err != nil {
		return nil, errors.New("Couldn't render post in HTML")
	}

	post := atom.Text{
		Type: "html",
		Body: *htmlPost,
	}

	nl := strings.Index(e.Post.Text, "\n")
	if nl < 0 && len(e.Post.Text) <= maxtitle {
		title = e.Post.Text
	} else if nl >= 0 && nl <= maxtitle {
		title = e.Post.Text[:nl]
	} else {
		title = fmt.Sprintf("%s...", e.Post.Text[:(maxtitle-3)])
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

func renderFeed(f *FeedMeta, p []pointapi.PostMeta) ([]byte, error) {
	var posts []*atom.Entry
	var timestamp atom.TimeStr

	for i := range p {
		entry, err := makeEntry(&p[i])
		if err != nil {
			return nil, err
		}
		posts = append(posts, entry)
	}

	if len(p) > 0 {
		timestamp = atom.Time(p[0].Post.Created)
	} else {
		timestamp = atom.Time(time.Now())
	}

	feed := atom.Feed{
		Title: f.Title,
		ID:    f.ID,
		Link: []atom.Link{
			atom.Link{
				Rel:  "alternate",
				Href: f.Href,
			},
		},
		Updated: timestamp,
		Entry:   posts,
	}

	return xml.Marshal(feed)
}
