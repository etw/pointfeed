package main

import (
	"bytes"
	"net/url"
	"regexp"
	"text/template"

	"github.com/russross/blackfriday"
)

const (
	mdHtmlFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
		blackfriday.HTML_NOREFERRER_LINKS |
		blackfriday.HTML_SMARTYPANTS_ANGLED_QUOTES

	mdExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS
)

type Text struct{}

type Html struct {
	*blackfriday.Html
}

var (
	mdRenderer = newRenderer(mdHtmlFlags, "", "")
	rdRenderer = newRenderer(mdHtmlFlags|blackfriday.HTML_COMPLETE_PAGE, "Point.im feed proxy", "")
	htmlEntity = regexp.MustCompilePOSIX(`&[a-z]{2,5};`)
)

func newRenderer(flags int, title string, css string) blackfriday.Renderer {
	var r = new(Html)

	r.Html = blackfriday.HtmlRenderer(flags, title, css).(*blackfriday.Html)

	return r
}

func (options *Html) closeTag() string {
	if options.GetFlags()&blackfriday.HTML_USE_XHTML != 0 {
		return " />"
	} else {
		return ">"
	}
}

func entityEscapeWithSkip(out *bytes.Buffer, src []byte, skipRanges [][]int) {
	var end int
	for _, rang := range skipRanges {
		template.HTMLEscape(out, src[end:rang[0]])
		out.Write(src[rang[0]:rang[1]])
		end = rang[1]
	}
	template.HTMLEscape(out, src[end:])
}

func (options *Html) RawHtmlTag(out *bytes.Buffer, text []byte) {
	template.HTMLEscape(out, text)
}

func (options *Html) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	var newlink []byte
	if u, err := url.Parse(string(link)); err != nil {
		newlink = link
	} else {
		if urlHttps(u) {
			newlink = []byte(u.String())
		} else {
			newlink = link
		}
	}

	out.WriteString("<a href=\"")
	template.HTMLEscape(out, newlink)
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		template.HTMLEscape(out, title)
	}

	out.WriteString("\" rel=\"noreferrer\" target=\"_blank\">")
	out.Write(content)
	out.WriteString("</a>")
	return
}

func (options *Html) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	var (
		img = new(bytes.Buffer)
	)

	img.WriteString("<img src=\"")
	template.HTMLEscape(img, link)
	img.WriteString("\" alt=\"")
	if len(alt) > 0 {
		template.HTMLEscape(img, alt)
	}
	if len(title) > 0 {
		img.WriteString("\" title=\"")
		template.HTMLEscape(img, title)
	}

	img.WriteByte('"')
	img.WriteString(options.closeTag())

	options.Link(out, link, title, img.Bytes())
	img.Reset()
}

func (options *Html) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	var (
		skipRanges = htmlEntity.FindAllIndex(link, -1)
	)

	if u, err := url.Parse(string(link)); err == nil {
		if p, ok := urlGelbooru(u); ok {
			options.Image(out, []byte(p.Sample), []byte(p.Tags), link)
			return
		}
		if p, ok := urlDanbooru(u); ok {
			options.Image(out, []byte(p.Sample), []byte(p.Tags), link)
			return
		}
		if urlPicture(u) {
			options.Image(out, link, link, link)
			return
		}
	}

	out.WriteString("<a href=\"")
	if kind == blackfriday.LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}

	entityEscapeWithSkip(out, link, skipRanges)

	out.WriteString("\" rel=\"noreferrer\" target=\"_blank\">")

	if bytes.HasPrefix(link, []byte("mailto://")) {
		template.HTMLEscape(out, link[len("mailto://"):])
	} else if bytes.HasPrefix(link, []byte("mailto:")) {
		template.HTMLEscape(out, link[len("mailto:"):])
	} else {
		entityEscapeWithSkip(out, link, skipRanges)
	}

	out.WriteString("</a>")
}
