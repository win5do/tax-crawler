package main

import (
	"fmt"
	errors2 "github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path"
	"time"
)

const (
	bot = "https://oapi.dingtalk.com/robot/send?access_token=8dd63b49541ef8b6183df4a96a6f28efa22521ed0b6c1a4ee961c48297b4cdb9"
)

var (
	flagBot      string
	flagLogLevel string
	flagCron     int
	flagRange    int
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "main",
		Long: `tax crawler`,
		RunE: func(cmd *cobra.Command, args []string) error {
			lvl, err := log.ParseLevel(flagLogLevel)
			if err != nil {
				return err
			}
			log.Infof("log level: %s", lvl)
			log.SetLevel(lvl)
			log.SetReportCaller(true)
			openDB(path.Join("/opt/data", "tax.db"))
			return run()
		},
	}

	rootCmd.Flags().StringVar(&flagBot, "bot", bot, "bot webhook addr")
	rootCmd.Flags().IntVar(&flagCron, "cron", 30, "job exec interval minutes")
	rootCmd.Flags().IntVar(&flagRange, "range", 24*60, "news post range minutes")
	rootCmd.Flags().StringVar(&flagLogLevel, "verbose", "info", "log level")

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("err: %+v", err)
		return
	}
}

func run() error {
	register(site_country_html, site_shanghai_html)

	return cronJob()
}

func cronJob() error {
	c := cron.New()

	_, err := c.AddFunc(fmt.Sprintf("@every %dm", flagCron), func() {
		err := crawler(time.Now()) // 记录当前时间点，防止 task 执行中取 now 因执行耗时不一致产生时间间隙
		if err != nil {
			log.Errorf("err: %+v", err)
			err := notify(applyMsg(fmt.Sprintf("err: %+v", err)))
			if err != nil {
				log.Errorf("err: %+v", err)
			}
			return
		}
	})
	if err != nil {
		return errors2.WithStack(err)
	}

	log.Info("start cronjob")
	c.Run()
	return nil
}

func crawler(timing time.Time) error {
	log.Infof("begin crawler: %s", timing.Format(time.RFC3339))

	for _, fn := range callbackList {
		news, err := fn()
		if err != nil {
			return errors2.WithStack(err)
		}
		log.Tracef("news: %+v", news)
		err = handleNews(news, timing)
		if err != nil {
			return errors2.WithStack(err)
		}
	}

	return nil
}
