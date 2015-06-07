package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"text/template"
	"regexp"

	point "github.com/etw/pointapi"

	"github.com/russross/blackfriday"
)

var secSites = []string{
	"google.com",
	"google.ru",
	"github.com",
	"point.im",
	"juick.com",
	"bnw.im",
	"danbooru.donmai.us",
	"safebooru.donmai.us",
	"chan.sankakucomplex.com",
	"yande.re",
}

var dbSites = map[string]bool{
	"danbooru.donmai.us":      true,
	"konachan.com":            false,
	"chan.sankakucomplex.com": true,
	"yande.re":                true,
}

var dbPost = regexp.MustCompilePOSIX("^/post/[0-9]+$")

func urlGelbooru(u *url.URL) (*string, bool) {
	if !(u.Scheme == "http" && u.Host == "gelbooru.com" && u.Path == "/index.php") {
		return nil, false
	}

	query := u.Query()
	if v, ok := query["page"]; !ok || !(v[0] == "post" && len(v) == 1) {
		return nil, false
	}
	if v, ok := query["s"]; !ok || !(v[0] == "view" && len(v) == 1) {
		return nil, false
	}
	if v, ok := query["id"]; !ok {
		return nil, false
	} else {
		if val, err := strconv.Atoi(v[0]); err != nil {
			panic(fmt.Sprintf("Failed to parse post id %s: %s", v[0], err))
			return nil, false
		} else {
			if p, err := apiset.Gelbooru.GetById(val); err != nil {
				panic(fmt.Sprintf("Failed to retrieve gelbooru pic %s: %s", v, err))
			} else {
				if tmp, err := url.Parse((*p).Sample); err != nil {
					panic(fmt.Sprintf("Failed to parse sample url %s: %s", (*p).Sample, err))
				} else {
					u = tmp
					return &(*p).Tags, true
				}
			}
		}
	}
}

func urlDanbooru(u *url.URL) (*string, bool) {
	if !(((u.Scheme == "http") || (u.Scheme == "https")) && isKey(dbSites, &u.Host)) {
		return nil, false
	}

	if !dbPost.MatchString(u.Path) {
		return nil, false
	}

	return nil, false
}

func urlHttps(u *url.URL) bool {
	if u.Scheme == "http" && isElem(secSites, &u.Host) {
		u.Scheme = "https"
		return true
	}
	return false
}

func formatFiles(out *bytes.Buffer, f []string) {
	var tmp bytes.Buffer
	for i, _ := range f {
		fmt.Fprintf(&tmp, "\n![%s](%s)\n", f[i], f[i])
		out.Write(blackfriday.MarkdownOptions(tmp.Bytes(), mdRenderer,
			blackfriday.Options{Extensions: mdExtensions}))
		tmp.Reset()
	}
}

func formatPost(out *bytes.Buffer, p []byte) {
	var escPost bytes.Buffer

	template.HTMLEscape(&escPost, p)
	defer escPost.Reset()

	out.Write(blackfriday.MarkdownOptions(escPost.Bytes(), mdRenderer,
		blackfriday.Options{Extensions: mdExtensions}))
}

func renderPost(out *bytes.Buffer, p *point.PostData) (e error) {
	defer func() {
		if r := recover(); r != nil {
			e = errors.New(fmt.Sprintf("\n%s", r))
		}
	}()

	formatPost(out, []byte(p.Text))
	formatFiles(out, p.Files)
	return nil
}
