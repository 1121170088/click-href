package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	re *regexp.Regexp
	MaxDepth int = 3
	visited map[string] bool = make(map[string] bool)
	mu sync.Mutex
	home string
	homeDomain string
	showHomeHref bool
)

func init()  {
	//p *string, name string, value string, usage string
	flag.StringVar(&home, "url", "", "root page")
	flag.IntVar(&MaxDepth, "md", 0, "max depth")
	flag.BoolVar(&showHomeHref, "hr", false, "show home href")
	flag.Parse()
	re = regexp.MustCompile(`href="([a-zA-Z0-9/].+?)"`)
}
func findHrefs(txt string) (hrefs []string)  {
	hrefstr := re.FindAllSubmatch([]byte(txt), -1)
	for _, e := range hrefstr {
		subMactch := e
		if len(subMactch) >= 2 {
			hrefs = append(hrefs, string(subMactch[1]))
		}
	}
	return
}

func doIt(url string, depth int, c chan bool)  {
	c <-true
	defer func() {
		<-c
		log.Printf("...")
	}()
	httpIdx := strings.Index(url, "http")
	if httpIdx != 0 {
		log.Fatal(url + "dosn't include http prefix")
	}
	var visitedV bool
	mu.Lock()
	visitedV = visited[url]
	mu.Unlock()
	if visitedV {
		return
	}
	dm := findDomain(url)
	if dm != homeDomain && depth >= MaxDepth {
		return
	}
	resp, err := http.Get(url)
	mu.Lock()
	visited[url] = true
	mu.Unlock()
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if depth == 0 {
		if dm == homeDomain {

		} else {
			depth += 1
		}
	} else {
		depth +=1
	}

	hrefs := findHrefs(string(bytes))
	for _, href := range hrefs {
		httpIdx = strings.Index(href, "http")
		if httpIdx != 0 {
			slashIdx := strings.Index(href, "/")
			if slashIdx != 0 {
				href = homeDomain + "/" + href
			} else {
				href = homeDomain + href
			}
		}

		go doIt(href, depth, c)

	}

}
func doit(home string, depth int,  c  chan bool)  {

	idx := strings.Index(home, "http")
	if idx != 0 {
		log.Fatal(home, depth)
	}
	mu.Lock()
	_, ok := visited[home]
	mu.Unlock()
	if  ok {
		log.Printf("visited %v, %d", home, depth)
		return
	}

	c <- true
	defer func() {
		<- c
		log.Printf("...")
	}()

	var resp *http.Response
	var err error
	dm := findDomain(home)
	if dm != homeDomain && depth > 1 {
		log.Fatal("weeeeeeeeeeeeeeeeee")
	}
	if dm != homeDomain {
		if depth >= MaxDepth {
			//log.Printf("skiped MaxDepth: %v, %d", home, depth)
			return
		} else {
			resp ,err = http.Get(home)
			depth +=1
			//log.Printf("0000000000 : %v, %d", home, depth)
		}
	} else {
		if depth !=0 {
			if depth >= MaxDepth {
				//log.Printf("skiped MaxDepth: %v, %d", home, depth)
				return
			}
			depth +=1
		}
		resp ,err = http.Get(home)
	}
	mu.Lock()
	visited[home] = true
	mu.Unlock()

	if err != nil {
		return
	}
	//log.Printf("cliked %v, %d", home, depth)
	defer resp.Body.Close()
	txt, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	hrefs := findHrefs(string(txt))
	for _, href := range hrefs {
		idx := strings.Index(href, "http")
		if idx != 0 {
			if depth == 0 {
				href = homeDomain + "/" +  href
				mu.Lock()
				_, ok := visited[href]
				mu.Unlock()
				if  ok {
					//log.Printf("visited %v, %d", home, depth)
					continue
				}
				go doit(href, depth, c)
			}
		} else {
			mu.Lock()
			_, ok := visited[href]
			mu.Unlock()
			if  ok {
				//log.Printf("visited %v, %d", home, depth)
				continue
			}
			go doit(href, depth, c)
		}

	}
}

func main()  {

	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	process := make(chan bool, 50)
	homeDomain = findDomain(home)
	log.Printf("home domain %s", homeDomain)
	//go doit(home, 0, process)
	go doIt(home, 0, process)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	lastProcess := time.Now().Unix()
	go func() {
		for {
			select {
			case <-sigCh:
				done<-true
				return
			case <-ticker.C:
				if time.Now().Unix() - lastProcess >= 30 {

					l := len(process)
					if l == 0 {
						done<-true
						return
					} else {
						log.Printf("len(process) => %d", l)
					}

				}


			}
		}
	}()

	<- done
	log.Printf("======================================")
	var total = 0
	for k, _ := range visited {
		if !showHomeHref && findDomain(k) == homeDomain {
			continue
		}
		total++
		log.Printf("%s", k)
	}
	log.Printf("total :%d", total)
}

func findDomain(href string) string  {
	slash1 := strings.Index(href, "/")
	len := len(href)
	prefix := href
	if slash1 != -1 {
		slash1 += 1
		if slash1 <= len -1 {
			slash2 := strings.Index(href[slash1:], "/")
			if slash2 != -1 {
				slash2 = slash1 + slash2 + 1
				if slash2 <= len -1 {
					slash3 := strings.Index(href[slash2:], "/")
					if slash3 != -1 {
						slash3 = slash2 + slash3 + 1
						prefix = href[:slash3 - 1]

					}

				}
			}
		}
	}
	return prefix
}