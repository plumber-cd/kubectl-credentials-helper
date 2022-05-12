//go:build windows
// +build windows

package keychain

import "github.com/danieljoos/wincred"

var (
	ErrorDuplicateItem = errors.New("Secret already existed")
	ErrorItemNotFound = errors.New("Secret was not found")
)

func CreateSecret(clusterName, clusterEndpoint, credentials string) error {
	g := wincred.NewGenericCredential(clusterEndpoint)
	g.UserName = clusterName
	g.CredentialBlob = []byte(credentials)
	g.Persist = wincred.PersistLocalMachine
	g.Attributes = []wincred.CredentialAttribute{{Keyword: "label", Value: AccessGroup}}

	return g.Write()
}

func DeleteCredentials(clusterEndpoint string) error {
	g, err := wincred.GetGenericCredential(clusterEndpoint)
	if err != nil {
		return err
	}
	if g == nil {
		return nil
	}
	for _, attr := range g.Attributes {
		if strings.Compare(attr.Keyword, "label") == 0 &&
			bytes.Compare(attr.Value, []byte(AccessGroup)) == 0 {

			return g.Delete()
		}
	}
	return ErrorItemNotFound
}

func GetSecret(clusterEndpoint string) (string, string, error) {
	g, err := wincred.GetGenericCredential(clusterEndpoint)
	if err != nil {
		return err
	}
	if g == nil {
		return ErrorItemNotFound
	}
	for _, attr := range g.Attributes {
		if strings.Compare(attr.Keyword, "label") == 0 &&
			bytes.Compare(attr.Value, []byte(AccessGroup)) == 0 {

			return return g.UserName, string(g.CredentialBlob), nil
		}
	}
	return ErrorItemNotFound
}
