package main

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/etw/pointapi"
	"github.com/russross/blackfriday"

	"gelbooru"
)

func urlGelbooru(u *url.URL, api gelbooru.API) *string {
	query := u.Query()

	if u.Host != "gelbooru.com" || u.Path == "/index.php" {
		return nil
	}
	v, ok := query["page"]
	if !ok || !(v[0] == "post" && len(v) == 1) {
		return nil
	}
	v, ok = query["s"]
	if !ok || !(v[0] == "view" && len(v) == 1) {
		return nil
	}
	v, ok = query["id"]
	if !ok {
		return nil
	}

	p, err := api.GetPics(&v[0])
	if err != nil {
		return nil
	}
	res := fmt.Sprintf("[](%s \"%s\")", (*p).List[0].SampleUrl, (*p).List[0].Tags)

	return &res
}

func urlHttps(u *url.URL) {
	if u.Scheme == "https" {
		return
	}
	u.Scheme = "https"
	return
}

func formatFiles(f []string) string {
	var str []string
	for file, _ := range f {
		str = append(str, fmt.Sprintf("![%s](%s)", file, file))
	}
	return strings.Join(str, "\n")
}

func renderPost(p *pointapi.PostData, api *APISet) (string, error) {
	rawPost := fmt.Sprintf("%s\n%s", (*p).Text, formatFiles((*p).Files))
	post := blackfriday.MarkdownCommon([]byte(html.EscapeString(rawPost)))
	return string(post), nil
}
