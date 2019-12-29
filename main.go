package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"sync"
)

func worker(domains chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for domain := range domains {
		response, err := http.Get(domain)
		if err != nil {
			log.Fatalln(err)
		}

		if err == nil {
			log.Println(fmt.Sprintf("[%s] Status %d", domain, response.StatusCode))
		} else {
			log.Println(fmt.Sprintf("[%s] Error %s", domain, err.Error()))
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var threadsCount int
	flag.IntVar(&threadsCount, "threads", 15, "Count of used threads (goroutines)")
	flag.Parse()

	domains := make(chan string, 2)
	for _, domain := range flag.Args() {
		u, err := url.ParseRequestURI(domain)
		if err != nil {
			var validURL = regexp.MustCompile(`^[a-zA-z]*:\/\/`)
			if !validURL.MatchString(domain) {
				domains <- fmt.Sprintf("http://%s", domain)
			} else {
				log.Println(fmt.Sprintf("[%s] Status 400 - Error %s", domain, err.Error()))
			}
		} else if u.Scheme == "" || u.Host == "" {
			log.Println(fmt.Sprintf("[%s] Status 400 - Must be an absolute URL", domain))
		} else if u.Scheme != "http" && u.Scheme != "https" {
			log.Println(fmt.Sprintf("[%s] Status 400 - Scheme must be HTTP or HTTPS", domain))
		} else {
			domains <- u.String()
		}
	}
	close(domains)

	wg := new(sync.WaitGroup)
	for i := 0; i < threadsCount; i++ {
		wg.Add(1)
		go worker(domains, wg)
	}
	wg.Wait()
}
