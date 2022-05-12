//go:build windows
// +build windows

package keychain

import (
	"bytes"
	"errors"
	"strings"

	"github.com/danieljoos/wincred"
)

var (
	ErrorDuplicateItem = errors.New("Secret already existed")
	ErrorItemNotFound  = wincred.ErrElementNotFound
)

func CreateSecret(clusterName, clusterEndpoint, credentials string) error {
	_, existing, err := GetSecret(clusterEndpoint)
	if err != ErrorItemNotFound || existing != "" {
		return ErrorDuplicateItem
	}

	g := wincred.NewGenericCredential(clusterEndpoint)
	g.UserName = clusterName
	g.CredentialBlob = []byte(credentials)
	g.Persist = wincred.PersistLocalMachine
	g.Attributes = []wincred.CredentialAttribute{{Keyword: "label", Value: []byte(AccessGroup)}}

	return g.Write()
}

func DeleteSecret(clusterEndpoint string) error {
	g, err := wincred.GetGenericCredential(clusterEndpoint)
	if err != nil && err != ErrorItemNotFound {
		return err
	}
	if g == nil {
		return nil
	}
	for _, attr := range g.Attributes {
		if strings.Compare(attr.Keyword, "label") == 0 &&
			bytes.Equal(attr.Value, []byte(AccessGroup)) {

			return g.Delete()
		}
	}
	return nil
}

func GetSecret(clusterEndpoint string) (string, string, error) {
	g, err := wincred.GetGenericCredential(clusterEndpoint)
	if err != nil {
		return "", "", err
	}
	if g == nil {
		return "", "", ErrorItemNotFound
	}
	for _, attr := range g.Attributes {
		if strings.Compare(attr.Keyword, "label") == 0 &&
			bytes.Equal(attr.Value, []byte(AccessGroup)) {

			return g.UserName, string(g.CredentialBlob), nil
		}
	}
	return "", "", ErrorItemNotFound
}
