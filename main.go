package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
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
	re = regexp.MustCompile(`href="(http.+?)"`)
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

func doit(home string, depth int,  c  chan bool)  {
	c <- true
	defer func() {
		<- c
		log.Printf("...")
	}()

	resp, err := http.Get(home)
	dm := findDomain(home)
	if dm != homeDomain {
		depth +=1
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	txt, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	hrefs := findHrefs(string(txt))
	for _, href := range hrefs {
		dm := findDomain(href)
		if dm != homeDomain {
			if depth >= MaxDepth {
				log.Printf("skiped MaxDepth: %v, %d", href, depth)
				continue
			}
		}

		//slash1 := strings.Index(href, "/")
		//len := len(href)
		prefix := href
		//if slash1 != -1 {
		//	slash1 += 1
		//	if slash1 <= len -1 {
		//		slash2 := strings.Index(href[slash1:], "/")
		//		if slash2 != -1 {
		//			slash2 = slash1 + slash2 + 1
		//			if slash2 <= len -1 {
		//				slash3 := strings.Index(href[slash2:], "/")
		//				if slash3 != -1 {
		//					slash3 = slash2 + slash3 + 1
		//					prefix = href[:slash3 - 1]
		//
		//				}
		//
		//
		//			}
		//		}
		//	}
		//}

		mu.Lock()
		_, ok := visited[prefix]
		mu.Unlock()
		if  ok {
			//log.Printf("skiped prefix: %v", prefix)
			continue
		}
		mu.Lock()
		visited[prefix] = true
		mu.Unlock()
		go doit(href, depth, c)
	}
}

func main()  {



	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	process := make(chan bool, 50)
	homeDomain = findDomain(home)
	log.Printf("home domain %s", homeDomain)
	go doit(home, 0, process)

	lastProcess := time.Now().Unix()
	go func() {
		for {
			select {
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