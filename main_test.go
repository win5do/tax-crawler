package main

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSite1(t *testing.T) {
	site_shanghai()
}

func TestSite2(t *testing.T) {
	site_country()
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
