package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/etw/gelbooru"
	"github.com/etw/pointapi"
	"golang.org/x/net/proxy"
)

type APISet struct {
	Point    *pointapi.API
	Gelbooru *gelbooru.API
}

func main() {
	var purl, host, port string
	var auth string

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

	apiset := &APISet{
		Point:    pointapi.New(&client, &auth),
		Gelbooru: gelbooru.New(&client),
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/feed/all", allHandler(apiset))
	http.HandleFunc("/feed/tags", tagsHandler(apiset))

	bind := fmt.Sprintf("%s:%s", host, port)
	log.Printf("[INFO] Listening on %s\n", bind)
	log.Fatalln(http.ListenAndServe(bind, nil))
}
