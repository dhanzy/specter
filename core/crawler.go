package core

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type CrawlerVisitedOptions struct {
	frameworkDetected bool
	links             map[string]struct{}
}

type Crawler struct {
	client *http.Client
	opts   CrawlerOptions

	visited       map[string]map[string]struct{}
	hostDepth     map[string]int
	hostFramework map[string]*DetectionResult
}

type CrawlerResponse struct {
	Target    Target
	Detection DetectionResult
}

type CrawlerOptions struct {
	MaxDepth         int
	UserAgent        string
	BlacklistDomains []string
	QueueSize        int
	Workers          int
}

func NewCrawler(client *http.Client, options CrawlerOptions) *Crawler {
	return &Crawler{
		client:        client,
		opts:          options,
		visited:       make(map[string]map[string]struct{}),
		hostDepth:     make(map[string]int),
		hostFramework: make(map[string]*DetectionResult),
	}
}

func (c *Crawler) Crawl(seeds []Target, targets chan<- DetectionResult) error {
	defer close(targets)
	// queue := make([]Target, 0, c.opts.QueueSize)
	queue := make(chan Target, c.opts.QueueSize)
	var wg sync.WaitGroup
	// queue = append(queue, seeds...)

	detector := NewFrameworkDetector()

	for i := 0; i < c.opts.Workers; i++ {
		go func() {
			for target := range queue {
				c.handle(target, queue, detector, targets, &wg)
			}
		}()
	}

	for _, seed := range seeds {
		wg.Add(1)
		queue <- seed
	}

	go func() {
		wg.Wait()
		close(queue)
	}()

	wg.Wait()

	return nil
}

func (c *Crawler) handle(target Target, queue chan<- Target, detector *FrameworkDetector, targets chan<- DetectionResult, wg *sync.WaitGroup) {
	defer wg.Done()

	if c.hostDepthExceeded(target.URL) {
		return
	}

	fmt.Printf("Crawling %s\n", target.URL.String())

	links, err := c.fetchLinks(target.URL, detector, targets)
	if err != nil {
		return
	}

	for _, link := range links {
		// Check if the link is in the same domain and not blacklisted in c.cfg.BlacklistDomains
		if c.opts.BlacklistDomains != nil {
			skip := false
			for _, blacklist_domain := range c.opts.BlacklistDomains {
				domain, err := c.GetDomainFromURL(link)
				if err != nil {
					skip = true
					break
				}
				if domain == blacklist_domain {
					skip = true
					break
				}

			}
			if skip {
				// fmt.Printf("Skipping link %s\n", link)
				continue
			}
		}

		// Checks if a link has already been visited, if not mark it as visited and add to the queue
		if !c.markVisited(link) {
			// fmt.Printf("Already visited %s\n", link)
			continue
		}

		wg.Add(1)

		// *queue = append(*queue, Target{URL: link})
		queue <- Target{URL: link}
	}
	return
}

func (c *Crawler) updateDepth(link *url.URL) {
	hostname := link.Hostname()
	total, ok := c.hostDepth[hostname]

	if !ok {
		c.hostDepth[hostname] = 0
	}
	c.hostDepth[hostname] = total + 1

}

func (c *Crawler) hostDepthExceeded(link *url.URL) bool {
	hostname := link.Hostname()
	total, ok := c.hostDepth[hostname]

	if !ok {
		return false
	}

	if total < c.opts.MaxDepth {
		return false
	}
	return true

}

func (c *Crawler) fetchLinks(targetURL *url.URL, detector *FrameworkDetector, targets chan<- DetectionResult) ([]*url.URL, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.opts.UserAgent != "" {
		req.Header.Set("User-Agent", c.opts.UserAgent)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var detectionResult *DetectionResult
	var ok bool

	hostname := targetURL.Hostname()
	if detectionResult, ok = c.hostFramework[hostname]; !ok {
		detectionResult, err = detector.Detect(resp)
		if err == nil {
			c.hostFramework[hostname] = detectionResult
		}
	}

	// *targets = append(*targets, *detectionResult)
	targets <- *detectionResult

	nodes, err := html.Parse(resp.Body)

	if err != nil {
		return nil, err
	}

	collected := []*url.URL{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					resolved, err := url.Parse(attr.Val)
					if err == nil && resolved.Scheme != "" && resolved.Host != "" {
						collected = append(collected, resolved)
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(nodes)
	c.updateDepth(targetURL)
	return collected, nil
}

func (c *Crawler) markVisited(u *url.URL) bool {
	key := u.String()

	domain, err := c.GetDomainFromURL(u)
	if err != nil {
		return false
	}

	if _, ok := c.visited[domain]; !ok {
		c.visited[domain] = make(map[string]struct{})
	}

	if _, ok := c.visited[domain][key]; ok {
		return false
	}

	c.visited[domain][key] = struct{}{}
	return true

}

func (c *Crawler) GetDomainFromURL(u *url.URL) (string, error) {
	parsedURL, err := url.Parse(u.String())
	if err != nil {
		return "", fmt.Errorf("error parsing  URL: %v", err)
	}

	hostname := parsedURL.Hostname()

	domain := strings.TrimPrefix(hostname, "www.")
	return domain, nil
}
