package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	errors2 "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Callback func() ([]News, error)

var callbackList []Callback

func register(cb ...Callback) {
	callbackList = append(callbackList, cb...)
}

func site_shanghai_html() ([]News, error) {
	var r []News

	domain := "http://shanghai.chinatax.gov.cn/zcfw/"

	c := NewCollector()

	c.OnHTML(".content li", func(e *colly.HTMLElement) {
		date, err := time.Parse("2006-01-02", e.ChildText("span"))
		if err != nil {
			log.Warnf("time parse err: %s", err)
			return
		}

		r = append(r, News{
			Subject:  "上海税务局",
			Title:    e.ChildAttr("a[title]", "title"),
			Keywords: "",
			Url:      domain + e.ChildAttr("a[href]", "href")[2:], // `./xx/yy.html`
			Date:     NewDate(date),
		})
	})

	err := c.Visit(domain)
	if err != nil {
		return nil, errors2.WithStack(err)
	}

	log.Debugf("news len: %d", len(r))
	return r, nil
}

func site_country_html() ([]News, error) {
	var r []News

	domain := "http://www.chinatax.gov.cn"
	path := "/chinatax/n810341/index.html"

	c := NewCollector()

	c.OnHTML(".zxwj_bottom li", func(e *colly.HTMLElement) {
		date, err := time.Parse("[01-02]", e.ChildText("a > span"))
		if err != nil {
			log.Warnf("time parse err: %s", err)
			return
		}

		if date.Month() > time.Now().Month() {
			date = date.AddDate(time.Now().Year()-1, 0, 0)
		} else {
			date = date.AddDate(time.Now().Year(), 0, 0)
		}

		r = append(r, News{
			Subject:  "国家税务局",
			Title:    e.ChildAttr("a[title]", "title"),
			Keywords: "",
			Url:      domain + e.ChildAttr("a[href]", "href"),
			Date:     NewDate(date),
		})
	})

	// 图解税收
	c.OnHTML(".tjss li", func(e *colly.HTMLElement) {
		url := domain + e.ChildAttr("a[href]", "href")

		var date time.Time
		cc := NewCollector()
		cc.OnHTML(".share_l", func(e *colly.HTMLElement) {
			var err error
			date, err = time.Parse("2006年01月02日", e.ChildText("span:first-child"))
			if err != nil {
				log.Warnf("time parse err: %s", err)
				return
			}
		})

		err := cc.Visit(url)
		if err != nil {
			log.Errorf("err: %+v", errors2.WithStack(err))
			return
		}

		r = append(r, News{
			Subject:  "国家税务局",
			Title:    e.ChildAttr("a[title]", "title"),
			Keywords: "",
			Url:      url,
			Date:     NewDate(date),
		})
	})

	err := c.Visit(domain + path)
	if err != nil {
		return nil, errors2.WithStack(err)
	}

	log.Debugf("news len: %d", len(r))
	return r, nil
}

func site_country_api() ([]News, error) {
	var r []News

	domain := "http://www.chinatax.gov.cn/api/query?siteCode=bm29000fgk&tab=all&key=9A9C42392D397C5CA6C1BF07E2E0AA6F"
	var domainCookie []*http.Cookie

	// ---> get cookie
	c := NewCollector()
	c.OnResponse(func(res *colly.Response) {
		cookies := c.Cookies(domain)
		if len(cookies) == 0 {
			return
		}
		domainCookie = cookies
	})
	err := c.Visit(domain)
	if err != nil {
		return nil, errors2.WithStack(err)
	}

	// --- get data
	c2 := NewCollector()
	err = c2.SetCookies(domain, domainCookie)
	if err != nil {
		return nil, errors2.WithStack(err)
	}
	c2.OnResponse(func(res *colly.Response) {
		log.Tracef("res: %s", string(res.Body))
		js := gjson.Parse(string(res.Body))
		for _, v := range js.Get("resultList").Array() {

			date, err := time.Parse("2006-01-02", v.Get("publishTime").String()[:10])
			if err != nil {
				log.Warnf("time parse err: %s", err)
				return
			}

			r = append(r, News{
				Subject:  "国家税务局",
				Title:    v.Get("title").String(),
				Keywords: v.Get("customHs.C6").String(),
				Url:      v.Get("url").String(),
				Date:     NewDate(date),
			})
		}

		log.Debugf("news len: %d", len(r))
	})

	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	header.Set("User-Agent", userAgent())

	err = c2.Request(
		http.MethodPost,
		domain,
		strings.NewReader("timeOption=0&page=1&pageSize=10&keyPlace=1&sort=dateDesc&qt=*"),
		nil,
		header,
	)
	if err != nil {
		return nil, errors2.WithStack(err)
	}
	return r, nil
}

type (
	News struct {
		Subject  string
		Title    string
		Url      string
		Keywords string
		Date     Date
	}

	Date struct {
		time.Time
	}
)

func NewDate(t time.Time) Date {
	return Date{
		t,
	}
}

func (s Date) String() string {
	return s.Format("2006-01-02")
}

func handleNews(news []News, timing time.Time) error {
	for _, v := range news {
		// date 只有日期，没有精确到时分秒，取最近一天的
		if v.Date.Before(timing.Add(-time.Duration(flagRange) * time.Minute)) {
			continue
		}

		key := hashKey(v.Url)
		_, err := Find(key)
		if err != nil {
			if !errors2.Is(err, Err_not_found) {
				return errors2.WithStack(err)
			}
		} else {
			// found, skip
			log.Infof("already notify: %s", v.Title)
			continue
		}

		msg, err := applyNews(v)
		if err != nil {
			return errors2.WithStack(err)
		}

		err = notify(msg)
		if err != nil {
			return errors2.WithStack(err)
		}

		err = Save(key, "1")
		if err != nil {
			return errors2.WithStack(err)
		}
	}
	return nil
}
