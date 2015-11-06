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

	"github.com/golang/groupcache/lru"
	"github.com/golang/groupcache/singleflight"
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

	doGroup *singleflight.Group
	pCache  *lru.Cache
	stats   Stats
)

func main() {
	var (
		purl string // Proxy URL
		host string // Listen host
		port string // Listen port
		auth string // Authentication token
		ddir string // Data directory
		pcsz int    //Size of posts cache

		gbauth *booru.GbAuth // Gelbooru auth data

		socks proxy.Dialer
	)

	flag.StringVar(&purl, "proxy", "", "SOCKS5 proxy URI (e.g socks5://localhost:9050/)")
	flag.StringVar(&host, "host", "localhost", "Interface to listen")
	flag.StringVar(&port, "port", "8000", "Port to listen")
	flag.StringVar(&auth, "auth", "", "Authentication token")
	flag.StringVar(&ddir, "data", ".", "Data directory")
	flag.StringVar(&ddir, "gbuser", ".", "Gelbooru user id")
	flag.StringVar(&ddir, "gbhash", ".", "Gelbooru password hash")
	flag.IntVar(&loglvl, "loglevel", INFO, "Logging level [0-4]")
	flag.IntVar(&pcsz, "pcachesz", pCacheSize, "Size of posts cache")
	flag.Parse()

	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("PORT")) > 0 {
		host = os.Getenv("HOST")
		port = os.Getenv("PORT")
		logger(INFO, "Got host:port from environment variables")
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

	if len(os.Getenv("GB_USER")) > 0 && len(os.Getenv("GB_HASH")) > 0 {
		gbauth = &booru.GbAuth{
			User: os.Getenv("GB_USER"),
			Hash: os.Getenv("GB_HASH"),
		}
		logger(INFO, "Got gelbooru auth from environment")
	}

	apiset = &APISet{
		Point:    point.New(&client, point.POINTAPI, &auth),
		Gelbooru: booru.NewGb(&client, booru.GELBOORU, gbauth),
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

	doGroup = new(singleflight.Group)
	pCache = lru.New(pcsz)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/stats/", statsHandler)
	http.HandleFunc("/stats/cache", cacheHandler)
	http.HandleFunc("/stats/media", mediaHandler)
	http.HandleFunc("/feed/all", allHandler)
	http.HandleFunc("/feed/tags", tagsHandler)

	bind := fmt.Sprintf("%s:%s", host, port)
	logger(INFO, fmt.Sprintf("Listening on %s", bind))
	if err := http.ListenAndServe(bind, nil); err != nil {
		logger(FATAL, err)
	}
}
