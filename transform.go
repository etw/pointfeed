package main

import (
	"bytes"
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

var (
	mdRenderer = blackfriday.HtmlRenderer(mdHtmlFlags, "", "")
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
			logger(WARN, fmt.Sprintf("Failed to parse post id %s: %s", v[0], err))
			return nil
		} else {
			if p, err := apiset.Gelbooru.GetByIdRaw(val); err != nil {
				logger(WARN, fmt.Sprintf("Failed to retrieve gelbooru pic %s: %s", v, err))
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

func formatFiles(out *bytes.Buffer, f []string) {
	for i, _ := range f {
		var tmp []byte
		tmp = append(tmp, []byte("![")...)
		tmp = append(tmp, []byte(f[i])...)
		tmp = append(tmp, []byte("](")...)
		tmp = append(tmp, []byte(f[i])...)
		tmp = append(tmp, []byte(")\n")...)
		out.Write(blackfriday.MarkdownOptions(tmp, mdRenderer,
			blackfriday.Options{Extensions: mdExtensions}))
	}
}

func formatPost(out *bytes.Buffer, p []byte) {
	var escPost bytes.Buffer

	template.HTMLEscape(&escPost, p)
	defer escPost.Reset()

	out.Write(blackfriday.MarkdownOptions(escPost.Bytes(), mdRenderer,
		blackfriday.Options{Extensions: mdExtensions}))
}

func renderPost(out *bytes.Buffer, p *pointapi.PostData) error {
	formatPost(out, []byte(p.Text))
	formatFiles(out, p.Files)
	return nil
}
