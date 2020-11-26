package main

import (
	"github.com/gocolly/colly/v2"
	"time"
)

func NewCollector() *colly.Collector {
	c := colly.NewCollector()
	c.UserAgent = userAgent()
	c.SetRequestTimeout(10 * time.Second)

	return c
}

func userAgent() string {
	return `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36 Edg/86.0.622.63`
}
