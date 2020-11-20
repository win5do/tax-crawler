package main

import (
	errors2 "github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	//eid, err := c.AddFunc("@daily", func() {
	_, err := c.AddFunc("* * * * *", func() {
		err := crawler()
		if err != nil {
			log.Errorf("err: %+v", err)
		}
	})
	if err != nil {
		return errors2.WithStack(err)
	}

	log.Info("start cronjob")
	c.Run()
	return nil
}

func crawler() error {
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
