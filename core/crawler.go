package core

import (
	"fmt"
	"net/http"
	"net/url"
)

type Crawler struct {
	client *http.Client
	opts   CrawlerOptions
}

type CrawlerOptions struct {
	MaxDepth         int
	UserAgent        string
	BlacklistDomains []string
}

func NewCrawler(client *http.Client, options CrawlerOptions) *Crawler {
	return &Crawler{client: client, opts: options}
}

func (c *Crawler) Crawl(target *url.URL) error {
	c.fetchLinks(target)
	return nil
}

func (c *Crawler) fetchLinks(url *url.URL) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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

	fmt.Printf("Fetched %s - Status: %d\n", url, resp.StatusCode)
	return []string{}, nil
}
