package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func init() {
	flagBot = bot
}

func TestApplyTpl(t *testing.T) {
	r, err := applyNews(News{
		"x",
		"x",
		"x",
		"x",
		NewDate(time.Now()),
	})
	require.NoError(t, err)
	t.Log(r)
}

func TestNotifyNews(t *testing.T) {
	err := notify(`{
         "msgtype": "markdown",
         "markdown": {
             "title":"CICD",
             "text": "# [x](x) \n> x \n\n x \n"
         }
        }`)
	require.NoError(t, err)
}

func TestNotifyErr(t *testing.T) {
	err := notify(applyMsg(fmt.Sprintf(`err: Get "http://xxx.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)\nmain.site_shanghai_html\n\t/go/src/app/crawler.go:47\nmain.crawler\n\t/go/src/app/main.go:79\nmain.cronJob.func1\n\t/go/src/app/main.go:61\ngithub.com/robfig/cron/v3.FuncJob.Run`)))
	require.NoError(t, err)
}
