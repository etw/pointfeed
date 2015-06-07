package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"text/template"

	booru "github.com/etw/gobooru"
	point "github.com/etw/pointapi"

	"github.com/russross/blackfriday"
)

const imgFmt = `<p><a href="%s" rel="noreferrer" target="_blank"><img src="%s" alt="%s" title="%s" /></a></p>`

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

var (
	dbPost = regexp.MustCompilePOSIX(`^/post/[0-9]+$`)
	imgExt = regexp.MustCompilePOSIX(`.(jpe?g|JPE?G|png|PNG|gif|GIF)$`)
	vidExt = regexp.MustCompilePOSIX(`.(flv|FLV|webm|WEBM|mp4|MP4)$`)
	audExt = regexp.MustCompilePOSIX(`.(mp3|MP3|m4a|M4A|ogg|OGG)$`)
)

func urlGelbooru(u *url.URL) (*booru.Post, bool) {
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
				return p, true
			}
		}
	}
}

func urlDanbooru(u *url.URL) (*booru.Post, bool) {
	if !(((u.Scheme == "http") || (u.Scheme == "https")) && isKey(dbSites, &u.Host)) {
		return nil, false
	}

	if !dbPost.MatchString(u.Path) {
		return nil, false
	}

	if dbSites[u.Host] {
		urlHttps(u)
	}

	return nil, false
}

func urlImage(u *url.URL) bool {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return false
	}

	if imgExt.MatchString(u.Path) {
		urlHttps(u)
		return true
	}

	return false
}

func urlVideo(u *url.URL) bool {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return false
	}

	if vidExt.MatchString(u.Path) {
		urlHttps(u)
		return true
	}

	return false
}

func urlYoutube(u *url.URL) bool {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return false
	}

	if !((u.Host == "youtube.com") || (u.Host == "www.youtube.com") ||
		(u.Host == "youtu.be")) || !(u.Path == "/watch") {
		return false
	}

	return false
}

func urlAudio(u *url.URL) bool {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return false
	}

	if audExt.MatchString(u.Path) {
		urlHttps(u)
		return true
	}

	return false
}

func urlHttps(u *url.URL) bool {
	if u.Scheme == "http" && isElem(secSites, &u.Host) {
		u.Scheme = "https"
		return true
	}
	return false
}

func formatFiles(out *bytes.Buffer, f []string) {
	for i, _ := range f {
		var l string
		if u, err := url.Parse(f[i]); err != nil {
			l = template.HTMLEscapeString(f[i])
		} else {
			urlHttps(u)
			l = template.HTMLEscapeString(u.String())
		}
		fmt.Fprintf(out, imgFmt, l, l, l, l)
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
