//go:build linux
// +build linux

package keychain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBig(t *testing.T) {
	err := CreateSecret("foo", "http://foo", "bar")
	require.NoError(t, err)
}
