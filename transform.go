package main

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/etw/pointapi"
	"github.com/russross/blackfriday"
)

func urlGelbooru(u *url.URL) *string {
	var (
		query = u.Query()
	)

	if !(u.Host == "gelbooru.com" && u.Path == "/index.php") {
		return nil
	}

	if v, ok := query["page"]; !ok || !(v[0] == "post" && len(v) == 1) {
		return nil
	}

	if v, ok := query["s"]; !ok || !(v[0] == "view" && len(v) == 1) {
		return nil
	}

	if v, ok := query["id"]; !ok {
		return nil
	} else {
		if val, err := strconv.Atoi(v[0]); err != nil {
			log.Printf("[WARN] Failed to parse post id %s: %s\n", v[0], err)
			return nil
		} else {
			if p, err := api.Gelbooru.GetByIdRaw(val); err != nil {
				log.Printf("[WARN] Failed to retrieve gelbooru pic %s: %s\n", v, err)
				return nil
			} else {
				res := fmt.Sprintf("[](%s \"%s\")", (*p).List[0].SampleUrl, (*p).List[0].Tags)
				return &res
			}
		}
	}
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
	for i, _ := range f {
		str = append(str, fmt.Sprintf("![%s](%s)", f[i], f[i]))
	}
	return strings.Join(str, "\n")
}

func renderPost(out *bytes.Buffer, p *pointapi.PostData) error {
	rawPost := fmt.Sprintf("%s\n%s", (*p).Text, formatFiles((*p).Files))
	out.Write(blackfriday.MarkdownCommon([]byte(html.EscapeString(rawPost))))
	return nil
}
