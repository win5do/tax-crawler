package main

import (
	"bytes"
	"fmt"
	"github.com/go-resty/resty/v2"
	errors2 "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
	"text/template"
)

const (
	newsStr = `{
 "msgtype": "markdown",
 "markdown": {
     "title":"税务 {{.Subject}}",
     "text": "# [{{.Title}}]({{.Url}}) \n> {{.Keywords}} \n\n {{.Date}} \n"
 }
}`

	msgStr = `{
 "msgtype": "markdown",
 "markdown": {
     "title":"税务",
     "text": "%s"
 }
}`
)

var newsTpl = template.Must(template.New("News").Parse(newsStr))

func applyNews(data News) (string, error) {
	r := bytes.NewBufferString("")
	err := newsTpl.Execute(r, data)
	if err != nil {
		return "", errors2.WithStack(err)
	}
	return r.String(), nil
}

func applyMsg(msg string) string {
	return fmt.Sprintf(msgStr, strings.ReplaceAll(msg, `"`, `'`))
}

type BotResult struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func notify(str string) error {
	log.Infof("notify msg: %s", str)

	botRes := &BotResult{}

	_, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(str).
		SetResult(botRes).
		Post(flagBot)
	if err != nil {
		return errors2.WithStack(err)
	}

	if botRes.Errcode != 0 {
		return errors2.New(botRes.Errmsg)
	}

	return nil
}
