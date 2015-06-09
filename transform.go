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

const (
	imgFmt = `<p><a href="%s" rel="noreferrer" target="_blank"><img src="%s" alt="%s" title="%s" /></a></p>`
	ytFmt  = `<p><iframe id="ytPlayer" type="text/html" width="640" height="390" src="https://www.youtube.com/embed/%s" frameborder="0"></iframe></p>`
	cbFmt  = `<p><iframe id="coubVideo" type="text/html" width="450" height="360" src="https://coub.com/embed/%s" frameborder="0"></iframe></p>`
)

var secSites = []*regexp.Regexp{
	regexp.MustCompilePOSIX(`^google\.(com|ru)$`),
	regexp.MustCompilePOSIX(`^github\.com$`),
	regexp.MustCompilePOSIX(`^(i\.)?imgur\.com$`),
	regexp.MustCompilePOSIX(`^(i\.)?point\.im$`),
	regexp.MustCompilePOSIX(`^juick\.com$`),
	regexp.MustCompilePOSIX(`^bnw\.im$`),
	regexp.MustCompilePOSIX(`^(dan|safe)booru\.donmai\.us$`),
	regexp.MustCompilePOSIX(`^(chan\.)?sankakucomplex\.com$`),
	regexp.MustCompilePOSIX(`^yande\.re$`),
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
	cbPath = regexp.MustCompilePOSIX(`^/view/([[:alnum:]]+)$`)
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

func urlYoutube(u *url.URL) (*string, bool) {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return nil, false
	}

	query := u.Query()
	if ((u.Host == "youtube.com") || (u.Host == "www.youtube.com")) &&
		(u.Path == "/watch") {
		if v, ok := query["v"]; !ok {
			return nil, false
		} else {
			return &v[0], true
		}
	} else if u.Host == "youtu.be" {
		var id = u.Path[1:]
		return &id, true
	}

	return nil, false
}
func urlCoub(u *url.URL) (*string, bool) {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return nil, false
	}

	if (u.Host == "coub.com") && cbPath.MatchString(u.Path) {
		var ids = cbPath.FindStringSubmatch(u.Path)
		if len(ids) == 2 {
			return &ids[1], true
		}
	}

	return nil, false
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
	if u.Scheme == "http" && isMatching(secSites, &u.Host) {
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
