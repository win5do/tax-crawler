package main

import (
	"bytes"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly/v2"
	errors2 "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"net/url"
	"text/template"
	"time"
)

const (
	bot = "https://oapi.dingtalk.com/robot/send?access_token=4886c4e680073688ae1e2e247743c54a00c33bbf846f06c2cbc276eb91bc48d0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "main",
		Long: `tax crawler`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(log.DebugLevel)
			log.SetReportCaller(true)
			return run()
		},
	}
	//rootCmd.Flags().StringVar(&flagRepo, "repo", "", "git clone http url")

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("err: %+v", err)
		return
	}

}

func run() error {
	register(site_country, site_shanghai)
	for _, fn := range callbackList {
		news, err := fn()
		if err != nil {
			return errors2.WithStack(err)
		}
		log.Debugf("news: %+v", news)
		err = handleNews(news)
		if err != nil {
			return errors2.WithStack(err)
		}
	}
	return nil
}

type Callback func() ([]News, error)

var callbackList []Callback

func register(cb ...Callback) {
	callbackList = append(callbackList, cb...)
}

func site_shanghai() ([]News, error) {
	var r []News

	c := colly.NewCollector()

	c.OnHTML("ul#zcfglist > li", func(e *colly.HTMLElement) {
		date, err := time.Parse("2006-01-02", e.ChildText(".time"))
		if err != nil {
			log.Warnf("time parse err: %s", err)
			return
		}

		r = append(r, News{
			Subject:  "上海税务局",
			Title:    e.ChildText(".title"),
			Keywords: e.ChildText(".wh"),
			Url:      e.ChildAttr("a[href]", "href"),
			Date:     NewDate(date),
		})
	})

	err := c.Visit("http://shanghai.chinatax.gov.cn/zcfw/zcfgk/")
	if err != nil {
		return nil, errors2.WithStack(err)
	}

	log.Debugf("news len: %d", len(r))
	return r, nil
}

func site_country() ([]News, error) {
	var r []News

	vals, err := url.ParseQuery("timeOption=0&page=1&pageSize=10&keyPlace=1&sort=dateDesc&qt=*")
	if err != nil {
		return nil, errors2.WithStack(err)
	}

	log.Debugf("%+v", vals)
	//return nil, nil

	res, err := resty.New().R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetFormData(map[string]string{
			"sort": "dateDesc",
		}).
		Post("http://www.chinatax.gov.cn/api/query?siteCode=bm29000fgk&tab=all&key=9A9C42392D397C5CA6C1BF07E2E0AA6F")
	if err != nil {
		return nil, errors2.WithStack(err)
	}

	log.Debugf("res: %s", res.String())
	js := gjson.Parse(string(res.Body()))
	for _, v := range js.Get("resultList").Array() {

		date, err := time.Parse("2006-01-02", v.Get("publishTime").String()[:10])
		if err != nil {
			return nil, errors2.WithStack(err)
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

func handleNews(news []News) error {
	for _, v := range news {
		if v.Date.Before(time.Now().Add(-30 * 24 * time.Hour)) {
			continue
		}

		msg, err := applyTpl(v)
		if err != nil {
			return errors2.WithStack(err)
		}

		err = notify(msg)
		if err != nil {
			return errors2.WithStack(err)
		}
	}
	return nil
}

const newsTpl = `{
 "msgtype": "markdown",
 "markdown": {
     "title":"CICD {{.Subject}}",
     "text": "# [{{.Title}}]({{.Url}}) \n> {{.Keywords}} \n\n {{.Date}} \n"
 }
}`

var goTpl = template.Must(template.New("News").Parse(newsTpl))

func applyTpl(data News) (string, error) {
	r := bytes.NewBufferString("")
	err := goTpl.Execute(r, data)
	if err != nil {
		return "", errors2.WithStack(err)
	}
	return r.String(), nil
}

type BotResult struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func notify(str string) error {
	log.Debugf("notify msg: %s", str)

	botRes := &BotResult{}

	_, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(str).
		SetResult(botRes).
		Post(bot)
	if err != nil {
		return errors2.WithStack(err)
	}

	if botRes.Errcode != 0 {
		return errors2.New(botRes.Errmsg)
	}

	return nil
}
