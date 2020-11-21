package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSave(t *testing.T) {
	e := "world"
	err := Save("hello", e)
	require.NoError(t, err)
	r, err := Find("hello")
	require.NoError(t, err)
	require.Equal(t, e, r)
}

func TestHashKey(t *testing.T) {
	i := "hello"
	a := hashKey(i)
	b := hashKey(i)
	t.Log(a)
	require.Equal(t, a, b)
}
