package main

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestMain(t *testing.M) {
	log.SetLevel(log.DebugLevel)
	openDB("tax.db")
	t.Run()
}
