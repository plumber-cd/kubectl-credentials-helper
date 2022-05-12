//go:build darwin
// +build darwin

package keychain

import (
	"fmt"

	"github.com/keybase/go-keychain"
)

var (
	ErrorDuplicateItem = keychain.ErrorDuplicateItem
	ErrorItemNotFound  = keychain.ErrorItemNotFound
)

func CreateSecret(clusterName, clusterEndpoint, credentials string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(Service)
	item.SetAccount(clusterEndpoint)
	item.SetLabel(clusterName)
	item.SetAccessGroup(AccessGroup)
	item.SetData([]byte(credentials))
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	return keychain.AddItem(item)
}

func DeleteSecret(clusterEndpoint string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(Service)
	item.SetAccessGroup(AccessGroup)
	item.SetAccount(clusterEndpoint)
	return keychain.DeleteItem(item)
}

func GetSecret(clusterEndpoint string) (string, string, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(Service)
	query.SetAccessGroup(AccessGroup)
	query.SetAccount(clusterEndpoint)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnAttributes(true)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", "", err
	} else if len(results) != 1 {
		return "", "", fmt.Errorf("Multiple secrets for %s", clusterEndpoint)
	}
	return results[0].Label, string(results[0].Data), nil
}
