package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

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
