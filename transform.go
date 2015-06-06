package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"text/template"

	"github.com/etw/pointapi"
	"github.com/russross/blackfriday"
)

const (
	mdHtmlFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	mdExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK
)
	point "github.com/etw/pointapi"

var (
	mdRenderer = blackfriday.HtmlRenderer(mdHtmlFlags, "", "")
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

func urlHttps(u *url.URL) bool {
	if u.Scheme == "http" && isElem(secSites, u.Host) {
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
			e = errors.New(fmt.Sprintf("\n%s",r))
		}
	}()

	formatPost(out, []byte(p.Text))
	formatFiles(out, p.Files)
	return nil
}
