package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"sync/atomic"
	"text/template"

	booru "github.com/etw/gobooru"
	point "github.com/etw/pointapi"

	"github.com/russross/blackfriday"
)

const (
	imgFmt = `<p><a href="%s" rel="noreferrer" target="_blank"><img src="%s" alt="%s" title="%s" /></a></p>`
	audFmt = `<p><audo src="%s" controls><a href="%s" rel="noreferrer" target="_blank">%s</a></audio></p>`
	vidFmt = `<p><video src="%s" controls><a href="%s" rel="noreferrer" target="_blank">%s</a></video></p>`
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
	regexp.MustCompilePOSIX(`^konachan\.com$`),
	regexp.MustCompilePOSIX(`^2ch\.hk$`),
}

var dbSites = map[string]bool{
	"danbooru.donmai.us":      true,
	"konachan.com":            true,
	"chan.sankakucomplex.com": true,
	"yande.re":                true,
}

var (
	dbPost = regexp.MustCompilePOSIX(`^/post/([[:digit:]]+)$`)
	imgExt = regexp.MustCompilePOSIX(`.(jpe?g|JPE?G|png|PNG|gif|GIF)$`)
	vidExt = regexp.MustCompilePOSIX(`.(flv|FLV|webm|WEBM|mp4|MP4)$`)
	audExt = regexp.MustCompilePOSIX(`.(mp3|MP3|m4a|M4A|ogg|OGG)$`)
	cbPath = regexp.MustCompilePOSIX(`^/view/([[:alnum:]]+)$`)
)

func mediaTrans(out *bytes.Buffer, link []byte) bool {
	if u, err := url.Parse(string(link)); err == nil {
		if p, ok := urlGelbooru(u); ok {
			logger(DEBUG, fmt.Sprintf("Found gelbooru link: %s", link))
			out.WriteString(fmt.Sprintf(imgFmt, link, p.Sample, link, p.Tags))
			return true
		}
		if p, ok := urlDanbooru(u); ok {
			logger(DEBUG, fmt.Sprintf("Found danbooru link: %s", link))
			out.WriteString(fmt.Sprintf(imgFmt, link, p.Sample, link, p.Tags))
			return true
		}
		if urlImage(u) {
			logger(DEBUG, fmt.Sprintf("Found image link: %s", link))
			out.WriteString(fmt.Sprintf(imgFmt, link, link, link, link))
			return true
		}
		if urlAudio(u) {
			logger(DEBUG, fmt.Sprintf("Found audio link: %s", link))
			out.WriteString(fmt.Sprintf(audFmt, link, link, link))
			return true
		}
		if urlVideo(u) {
			logger(DEBUG, fmt.Sprintf("Found video link: %s", link))
			out.WriteString(fmt.Sprintf(vidFmt, link, link, link))
			return true
		}
		if id, ok := urlYoutube(u); ok {
			logger(DEBUG, fmt.Sprintf("Found youtube link: %s", link))
			out.WriteString(fmt.Sprintf(ytFmt, *id))
			return true
		}
		if id, ok := urlCoub(u); ok {
			logger(DEBUG, fmt.Sprintf("Found coub link: %s", link))
			out.WriteString(fmt.Sprintf(cbFmt, *id))
			return true
		}
	}
	return false
}

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
			logger(ERROR, fmt.Sprintf("Failed to parse post id %s: %s", v[0], err))
			return nil, false
		} else {
			if p, err := apiset.Gelbooru.GetById(val); err != nil {
				logger(ERROR, fmt.Sprintf("Failed to retrieve gelbooru pic %s: %s", v, err))
				return nil, false
			} else {
				atomic.AddUint64(&stats.Media.Gelbooru, 1)
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

	if s, ok := dbSites[u.Host]; ok && s {
		urlHttps(u)
		atomic.AddUint64(&stats.Media.Danbooru, 1)
	}
	return nil, false
}

func urlImage(u *url.URL) bool {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return false
	}

	if imgExt.MatchString(u.Path) {
		urlHttps(u)
		atomic.AddUint64(&stats.Media.Image, 1)
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
		atomic.AddUint64(&stats.Media.Video, 1)
		return true
	}
	return false
}

func urlAudio(u *url.URL) bool {
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		return false
	}

	if audExt.MatchString(u.Path) {
		urlHttps(u)
		atomic.AddUint64(&stats.Media.Audio, 1)
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
			atomic.AddUint64(&stats.Media.Youtube, 1)
			return &v[0], true
		}
	} else if u.Host == "youtu.be" {
		var id = u.Path[1:]
		atomic.AddUint64(&stats.Media.Youtube, 1)
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
			atomic.AddUint64(&stats.Media.Coub, 1)
			return &ids[1], true
		}
	}
	return nil, false
}

func urlHttps(u *url.URL) bool {
	if u.Scheme == "http" && isMatching(secSites, &u.Host) {
		atomic.AddUint64(&stats.Media.Https, 1)
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
