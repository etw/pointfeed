package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"crypto/rand"
	"encoding/base64"

	"github.com/etw/pointapi"
)

type FeedMeta struct {
	Title string
	ID    string
	Href  string
}

func getRid() (string, error) {
	rid := make([]byte, 8)
	_, err := rand.Read(rid)
	if err != nil {
		log.Println("[Error] Couldn't generate request id: %s", err)
		return "", errors.New("Couldn't generate request id")
	}
	return base64.URLEncoding.EncodeToString(rid), nil
}

func resRender(res http.ResponseWriter, rid *string, feed FeedMeta, body *map[string]interface{}) {
	posts := (*body)["posts"].([]interface{})
	result, err := renderFeed(feed, posts)
	if err != nil {
		log.Printf("[ERROR] Failed to parse point response: %s\n", err)
		res.WriteHeader(500)
		return
	}
	res.Header().Set("Content-Type", "application/atom+xml")
	res.Header().Set("Request-Id", *rid)
	fmt.Fprintln(res, string(result))
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "https://github.com/etw/pointfeed/blob/develop/README.md")
}

func allHandler(api *pointapi.PointAPI) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		//		params := req.URL.Query()
		rid, err := getRid()
		if err != nil {
			res.WriteHeader(500)
			return
		}
		log.Printf("[INFO] {%s} %s %s\n", rid, req.Method, req.RequestURI)

		body, err := api.GetAll(0)
		if err != nil {
			log.Printf("[ERROR] {%s} Failed to get all posts: %s\n", rid, err)
			res.WriteHeader(500)
			return
		}

		feed := FeedMeta{
			Title: "All posts",
			ID:    "all",
			Href:  "https://point.im/all",
		}

		resRender(res, &rid, feed, &body)
	}
}

func tagsHandler(api *pointapi.PointAPI) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		tags := params["tag"]
		rid, err := getRid()
		if err != nil {
			res.WriteHeader(500)
			return
		}
		log.Printf("[INFO] {%s} %s %s\n", rid, req.Method, req.RequestURI)

		if len(tags) < 1 {
			log.Printf("[WARN] {%s} At least one tag is needed\n", rid)
			res.WriteHeader(400)
			fmt.Fprintln(res, "At least one tag is needed")
			return
		}

		body, err := api.GetTags(0, tags)
		if err != nil {
			log.Printf("[ERROR] {%s} Failed to get tagged posts: %s\n", rid, err)
			res.WriteHeader(500)
			return
		}

		feed := FeedMeta{
			Title: fmt.Sprintf("Tagged posts (%s)", strings.Join(tags, ", ")),
			ID:    fmt.Sprintf("tags:%s", strings.Join(tags, ",")),
			Href:  fmt.Sprintf("https://point.im/?tag=%s", strings.Join(tags, "&tag=")),
		}

		resRender(res, &rid, feed, &body)
	}
}
