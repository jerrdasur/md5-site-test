package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	defaultParallel = 10
	httpPrefix      = "http://"
	httpsPrefix     = "https://"
)

func fetcher(tasks <-chan string, wg *sync.WaitGroup) {
	errLogger := log.New(os.Stderr, "", 0)

	for url := range tasks {
		if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
			url = httpPrefix + url
		}

		resp, err := http.Get(url)
		if err != nil {
			errLogger.Println(err)
			continue
		}

		hash := md5.New()
		_, err = io.Copy(hash, resp.Body)
		if err != nil {
			errLogger.Println(err)
			continue
		}

		_ = resp.Body.Close()
		fmt.Println(url, hex.EncodeToString(hash.Sum(nil)))
	}

	wg.Done()
}

func run(urls []string, maxParallelRequests int) {
	if maxParallelRequests > len(urls) {
		maxParallelRequests = len(urls)
	} else if maxParallelRequests <= 0 {
		maxParallelRequests = defaultParallel
	}

	tasks := make(chan string, len(urls))

	for _, url := range urls {
		tasks <- url
	}
	close(tasks)

	var wg sync.WaitGroup
	for i := 0; i < maxParallelRequests; i++ {
		wg.Add(1)
		go fetcher(tasks, &wg)
	}

	wg.Wait()
}

func main() {
	var maxParallelRequests int
	flag.IntVar(&maxParallelRequests, "parallel", defaultParallel, "max number of requests that can be run in parallel")
	flag.Parse()

	run(flag.Args(), maxParallelRequests)
}
