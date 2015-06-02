package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/etw/gobooru"
	"github.com/etw/pointapi"
	"github.com/russross/blackfriday"
	"golang.org/x/net/proxy"

	_ "net/http/pprof"
)

const readme = "README.md"

type APISet struct {
	Point    *pointapi.API
	Gelbooru *gobooru.API
}

var (
	rmf []byte
	api *APISet
)

func main() {
	var (
		purl string // Proxy URL
		host string // Listen host
		port string // Listen port
		auth string // Authentication token
	)

	flag.StringVar(&purl, "proxy", "", "SOCKS5 proxy URI (e.g socks5://localhost:9050/)")
	flag.StringVar(&host, "host", "localhost", "Interface to listen")
	flag.StringVar(&port, "port", "8000", "Port to listen")
	flag.StringVar(&auth, "auth", "", "Authentication token")
	flag.Parse()

	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("PORT")) > 0 {
		host = os.Getenv("HOST")
		port = os.Getenv("PORT")
		log.Println("[INFO] Got host:port fron environment variables")
	}

	if len(os.Getenv("POINT_AUTH")) > 0 {
		auth = os.Getenv("POINT_AUTH")
		log.Println("[INFO] Got token fron environment variable")
	}

	proxyuri, err := url.Parse(purl)
	if err != nil {
		log.Fatalf("[FATAL] %s is invalid URI\n", purl)
	}

	socks, err := proxy.FromURL(proxyuri, proxy.Direct)
	if err != nil {
		log.Println("[WARN] Fallback to direct connection")
		socks = proxy.Direct
	} else {
		log.Printf("[INFO] Using proxy %s\n", purl)
	}

	trans := &http.Transport{
		Dial:               socks.Dial,
		DisableCompression: true,
	}

	client := http.Client{
		Transport: trans,
	}

	api = &APISet{
		Point:    pointapi.New(&client, &auth),
		Gelbooru: gobooru.New(&client, gobooru.GbFmt),
	}

	rmraw, err := ioutil.ReadFile(readme)
	if err != nil {
		log.Fatalf("[FATAL] Couldn't read %s\n", readme)
	}
	rmf = blackfriday.MarkdownCommon(rmraw)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/feed/all", allHandler)
	http.HandleFunc("/feed/tags", tagsHandler)

	bind := fmt.Sprintf("%s:%s", host, port)
	log.Printf("[INFO] Listening on %s\n", bind)
	log.Fatalln(http.ListenAndServe(bind, nil))
}
