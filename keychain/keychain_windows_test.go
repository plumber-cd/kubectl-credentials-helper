//go:build windows
// +build windows

package keychain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBig(t *testing.T) {
	err := CreateSecret("foo", "http://foo", "bar")
	require.NoError(t, err)

	err = CreateSecret("baz", "http://foo", "qwe")
	require.Error(t, err)
	require.Equal(t, ErrorDuplicateItem, err)

	username, address, err := GetSecret("http://foo")
	require.NoError(t, err)
	require.Equal(t, "foo", username)
	require.Equal(t, "bar", address)

	err = DeleteSecret("http://foo")
	require.NoError(t, err)
}
