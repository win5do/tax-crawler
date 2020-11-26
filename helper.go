package main

import (
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
	"time"
)

func NewCollector() *colly.Collector {
	c := colly.NewCollector()
	c.UserAgent = userAgent()
	c.SetRequestTimeout(10 * time.Second)

	addRetry(c, 1) // retry 1 times

	return c
}

func addRetry(c *colly.Collector, retryExpect uint) {
	var retryCount uint = 0

	c.OnError(func(r *colly.Response, err error) {
		if err != nil {
			log.Errorf("err: %+v", err)
		}

		if retryCount < retryExpect {
			retryCount++
			log.Infof("retry: %d", retryCount)
			err := r.Request.Retry()
			if err != nil {
				log.Errorf("err: %+v", err)
			}
		}
	})
}

func userAgent() string {
	return `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36 Edg/86.0.622.63`
}
