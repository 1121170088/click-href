package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var (
	re *regexp.Regexp
	MaxDepth int = 3
	visited map[string] bool = make(map[string] bool)
	mu sync.Mutex
	home string
)

func init()  {
	//p *string, name string, value string, usage string
	flag.StringVar(&home, "url", "", "root page")
	flag.IntVar(&MaxDepth, "md", 0, "max depth")
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
		if depth >= MaxDepth {
			continue
		}
		mu.Lock()
		_, ok := visited[href]
		mu.Unlock()
		if  ok {
			continue
		}
		mu.Lock()
		visited[href] = true
		mu.Unlock()
		go doit(href, depth + 1, c)
	}
}

func main()  {



	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	process := make(chan bool, 50)
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

	for k, _ := range visited {
		log.Printf("%s", k)
	}
}