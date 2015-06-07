package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	booru "github.com/etw/gobooru"
	point "github.com/etw/pointapi"

	"github.com/russross/blackfriday"
	"golang.org/x/net/proxy"

	_ "net/http/pprof"
)

type APISet struct {
	Point    *point.API
	Gelbooru *booru.GbAPI
	Danbooru *booru.DbAPI
}

var (
	readme []byte
	apiset *APISet
	loglvl int
)

func main() {
	var (
		purl string // Proxy URL
		host string // Listen host
		port string // Listen port
		auth string // Authentication token
		ddir string // Data directory

		socks proxy.Dialer
	)

	flag.StringVar(&purl, "proxy", "", "SOCKS5 proxy URI (e.g socks5://localhost:9050/)")
	flag.StringVar(&host, "host", "localhost", "Interface to listen")
	flag.StringVar(&port, "port", "8000", "Port to listen")
	flag.StringVar(&auth, "auth", "", "Authentication token")
	flag.StringVar(&ddir, "data", ".", "Data directory")
	flag.IntVar(&loglvl, "loglevel", INFO, "Logging level [0-4]")
	flag.Parse()

	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("PORT")) > 0 {
		host = os.Getenv("HOST")
		port = os.Getenv("PORT")
		logger(INFO, "Got host:port fron environment variables")
	}

	if len(os.Getenv("POINT_AUTH")) > 0 {
		auth = os.Getenv("POINT_AUTH")
		logger(INFO, "Got token from environment variable")
	}

	if proxyuri, err := url.Parse(purl); err != nil {
		logger(FATAL, fmt.Sprintf("%s is invalid URI", purl))
	} else {
		if socks, err = proxy.FromURL(proxyuri, proxy.Direct); err != nil {
			logger(WARN, "Fallback to direct connection")
			socks = proxy.Direct
		} else {
			logger(INFO, fmt.Sprintf("Using proxy %s", purl))
		}
	}

	trans := &http.Transport{
		Dial:               socks.Dial,
		DisableCompression: true,
	}

	client := http.Client{
		Transport: trans,
	}

	apiset = &APISet{
		Point:    point.New(&client, point.POINTAPI, &auth),
		Gelbooru: booru.NewGb(&client, booru.GELBOORU),
		Danbooru: booru.NewDb(&client, booru.DANBOORU),
	}

	if len(os.Getenv("DATA_DIR")) > 0 {
		ddir = os.Getenv("DATA_DIR")
		logger(INFO, "Got data dir from environment variable")
	}

	if rmraw, err := ioutil.ReadFile(fmt.Sprintf("%s/README.md", ddir)); err != nil {
		logger(FATAL, "Couldn't read README.md")
	} else {
		readme = blackfriday.MarkdownOptions(rmraw, rdRenderer,
			blackfriday.Options{Extensions: mdExtensions})
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/feed/all", allHandler)
	http.HandleFunc("/feed/tags", tagsHandler)

	bind := fmt.Sprintf("%s:%s", host, port)
	logger(INFO, fmt.Sprintf("Listening on %s", bind))
	if err := http.ListenAndServe(bind, nil); err != nil {
		logger(FATAL, err)
	}
}
