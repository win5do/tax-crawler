package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestSite_shanghai(t *testing.T) {
	r, err := site_shanghai()
	require.NoError(t, err)
	t.Logf("%+v", r)
}

func TestSite_country(t *testing.T) {
	r, err := site_country()
	require.NoError(t, err)
	t.Logf("%+v", r)
}

func TestApplyTpl(t *testing.T) {
	r, err := applyTpl(News{
		"x",
		"x",
		"x",
		"x",
		NewDate(time.Now()),
	})
	require.NoError(t, err)
	t.Log(r)
}

func TestNotify(t *testing.T) {
	err := notify(`{
         "msgtype": "markdown",
         "markdown": {
             "title":"CICD",
             "text": "# [x](x) \n> x \n\n x \n"
         }
        }`)
	require.NoError(t, err)
}
