package main

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"golang.org/x/tools/blog/atom"
)

func makeEntry(e map[string]interface{}) (*atom.Entry, error) {
	var title string

	author := e["post"].(map[string]interface{})["author"].(map[string]interface{})["login"].(string)

	person := atom.Person{
		Name: author,
		URI:  fmt.Sprintf("https://%s.point.im/", author),
	}

	post := atom.Text{
		Type: "text",
		Body: e["post"].(map[string]interface{})["text"].(string),
	}

	nl := strings.Index(post.Body, "\n")
	if nl < 0 {
		title = post.Body
	} else {
		title = post.Body[:nl]
	}

	entry := atom.Entry{
		Title: title,
		ID:    fmt.Sprintf("%.0f", e["uid"].(float64)),
		Link: []atom.Link{
			atom.Link{
				Rel:  "alternate",
				Href: fmt.Sprintf("https://point.im/%s", e["post"].(map[string]interface{})["id"].(string)),
			},
		},
		Published: atom.TimeStr(e["post"].(map[string]interface{})["created"].(string)),
		Updated:   atom.TimeStr(e["post"].(map[string]interface{})["created"].(string)),
		Author:    &person,
		Content:   &post,
	}
	return &entry, nil
}

func renderFeed(f FeedMeta, p []interface{}) ([]byte, error) {
	var posts []*atom.Entry
	var timestamp atom.TimeStr

	for i := range p {
		entry, err := makeEntry(p[i].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		posts = append(posts, entry)
	}

	if len(p) > 0 {
		timestamp = atom.TimeStr(p[0].(map[string]interface{})["post"].(map[string]interface{})["created"].(string))
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
