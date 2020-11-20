package main

import (
	"fmt"
	errors2 "github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

const (
	bot = "https://oapi.dingtalk.com/robot/send?access_token=4886c4e680073688ae1e2e247743c54a00c33bbf846f06c2cbc276eb91bc48d0"
)

var (
	flagBot      string
	flagLogLevel string
	flagCron     int
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

			log.SetLevel(lvl)
			log.SetReportCaller(true)
			return run()
		},
	}

	rootCmd.Flags().StringVar(&flagBot, "bot", bot, "bot webhook addr")
	rootCmd.Flags().IntVar(&flagCron, "cron", 30, "interval minutes")
	rootCmd.Flags().StringVar(&flagLogLevel, "log-level", "info", "log level")

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("err: %+v", err)
		return
	}
}

func run() error {
	register(site_country, site_shanghai)

	return cronJob()
}

func cronJob() error {
	c := cron.New()

	_, err := c.AddFunc(fmt.Sprintf("@every %dm", flagCron), func() {
		err := crawler(time.Now()) // 记录当前时间点，防止 task 执行中取 now 因执行耗时不一致产生时间间隙
		if err != nil {
			log.Errorf("err: %+v", err)
			_ = notify(applyMsg(fmt.Sprintf("err: %+v", err)))
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
	for _, fn := range callbackList {
		news, err := fn()
		if err != nil {
			return errors2.WithStack(err)
		}
		log.Debugf("news: %+v", news)
		err = handleNews(news, timing)
		if err != nil {
			return errors2.WithStack(err)
		}
	}

	return nil
}
