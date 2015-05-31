package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/proxy"

	"github.com/etw/pointapi"
)

func main() {
	var purl string
	var host = os.Getenv("HOST")
	var port = os.Getenv("PORT")

	flag.StringVar(&purl, "proxy", "", "SOCKS5 proxy URI (e.g socks5://localhost:9050/)")
	flag.StringVar(&host, "host", "localhost", "Interface to listen")
	flag.StringVar(&port, "port", "8000", "Port to listen")
	flag.Parse()

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

	api := pointapi.New(&client, nil)

	http.HandleFunc("/feed/all", allHandler(api))
	http.HandleFunc("/feed/tags", tagsHandler(api))

	log.Fatalln(http.ListenAndServe(fmt.Sprint(host, ":", port), nil))
}




