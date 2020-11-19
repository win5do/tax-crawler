package main

import (
	"bytes"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly/v2"
	errors2 "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"text/template"
	"time"
	"github.com/tidwall/gjson"
)

const (
	bot = "https://oapi.dingtalk.com/robot/send?access_token=4886c4e680073688ae1e2e247743c54a00c33bbf846f06c2cbc276eb91bc48d0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "main",
		Long: `tax crawler`,
		Run: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.DebugLevel)
			log.SetReportCaller(true)
			err := run()
			if err != nil {
				log.Debugf("err: %+v", err)
				return
			}
		},
	}
	//rootCmd.Flags().StringVar(&flagRepo, "repo", "", "git clone http url")

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

}

func run() error {
	register(site_shanghai, site_country)

	for _, fn := range callbackList {
		news, err := fn()
		if err != nil {
			return errors2.WithStack(err)
		}
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
	return r, nil
}

func site_country() ([]News, error) {
	var r []News

	c := colly.NewCollector()
	c.OnResponse(func(res *colly.Response) {

		gjson.Get

		mp := make(map[string]interface{})
		err := json.Unmarshal(res.Body, &mp)
		if err != nil {
			log.Warnf("json unmarshal err: %s", err)
			return
		}

		list := mp["resultList"].([]interface{})

		var r []News
		for _, v := range list {
			obj := v.(map[string]interface{})

			date, err := time.Parse("2006-01-02", obj["publishTime"].(string)[:10])
			if err != nil {
				log.Warnf("time parse err: %s", err)
				return
			}

			r = append(r, News{
				Subject:  "国家税务局",
				Title:    obj["title"].(string),
				Keywords: obj["customHs"].(map[string]interface{})["C6"].(string),
				Url:      obj["url"].(string),
				Date:     NewDate(date),
			})
		}

		log.Infof("news: %+v", r)
	})

	err := c.Visit("http://www.chinatax.gov.cn/api/query?siteCode=bm29000fgk&tab=all&key=9A9C42392D397C5CA6C1BF07E2E0AA6F")
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

func handleNews(news []News) error {
	for _, v := range news {
		if v.Date.Before(time.Now().Add(-24 * time.Hour)) {
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
     "title":"{{.Subject}}",
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
