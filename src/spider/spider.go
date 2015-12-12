package spider

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

import (
	l4g "code.google.com/p/log4go"
	"golang.org/x/net/html"
)

import (
	conf "github.com/wusuopubupt/go_spider/src/conf"
)

type Spider struct {
	config   conf.SpiderCfg
	maxDepth config.maxDepth
}

// abnormal exit
func AbnormalExit() {
	// http://stackoverflow.com/questions/14252766/abnormal-behavior-of-log4go
	// adding a time.Sleep(time.Second) to the end of the code snippeet will cause the log content flush
	time.Sleep(time.Second)
	os.Exit(1)
}

// get href attribute from a Token
func GetHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
		l4g.Debug("get href: %s", href)
	}
	// 空的return默认返回ok, href
	return
}

// Extract all http** links from a given webpage
func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()
	if err != nil {
		l4g.Error("Failed to crawl %s, err[%s]", url, err)
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()
			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			// Extract the href value, if there is one
			ok, url := GetHref(t)
			if !ok {
				continue
			}
			// Make sure the url begines in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

/**
* @brief 爬取url
* @param seedUrls 种子urls数组
*
 */

func GetUrls(seedUrls []string) {
	// Channels
	/*
		c := make(chan bool) //创建一个无缓冲的bool型Channel
		c <- x //向一个Channel发送一个值
		<- c //从一个Channel中接收一个值
		x = <- c //从Channel c接收一个值并将其存储到x中
		x, ok = <- c //从Channel接收一个值，如果channel关闭了或没有数据，那么ok将被置为false
	*/
	chUrls := make(chan string)
	chFinished := make(chan bool)
	foundUrls := make(map[string]bool)

	// Kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		// 监听 IO 操作
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	l4g.Info("Found %d unique urls", len(foundUrls))
	for url, _ := range foundUrls {
		l4g.Info(" - " + url)
	}
	// close channel
	//chUrls <- "www.baidu.com"
	//url := <-chUrls
	close(chUrls)
}

// Crawler struct
type Spider struct {
	outputDir     string
	crawlInterval int
	crawlTimeout  int
	targetUrl     string
}

// one job
type Job struct {
	url   string
	depth int
}

// job queue
type JobQueue struct {
	url   chan string
	depth chan int
}

// get job from jobQueue
func (s *Spider) getJob(jobs *JobQueue) (job *Job) {
	url := <-jobs.url
	depth := <-jobs.depth
	return Job{url, depth}
}

// add job to jobQueue
func (s *Spider) addJob(jobs *JobQueue, job *Job) {
	jobs.url <- job.url
	jobs.depth <- job.depth
}

// crawl url
// which do current job and add new jobs to job queue
func (s *Spider) crawl(jobs *JobQueue) {

	// 抓取间隔控制
	time.Sleep(time.Duration(c.crawlInterval) * time.Second)
}

// new spider
func NewSpider(config conf.SpiderCfg) *Spider {
	s := new(Spider)
	s.spider = spider
	s.outDir = config.OutputDirectory
	s.crawlInterval = config.CrawlInterval
	s.crawlTimeout = config.CrawlTimeout

	return s
}

// 开启threandCount个spider goroutine,等待通道中的任务到达
func Start(seedUrl []string, spidercfg conf.SpiderCfg) {
	// 初始化任务队列
	jobs := new(JobQueue)
	for _, url := range seedUrls {
		jobs.url <- url
		jobs.depth <- 0
	}
	// 一个while(1)的循环，直到channel通知任务结束
	for {
		// 创建threadCount个工作goroutine
		for i := 0; i < spidercfg.ThreadCount; i++ {
			s := NewSpider(spidercfg)
			go s.crawl(jobs)
		}
	}
}