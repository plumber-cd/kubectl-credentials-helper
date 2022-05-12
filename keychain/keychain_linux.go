//go:build linux
// +build linux

package keychain

import (
	"errors"

	"github.com/keybase/dbus"
	"github.com/keybase/go-keychain/secretservice"
)

var (
	ErrorDuplicateItem = errors.New("Secret already existed")
	ErrorItemNotFound  = errors.New("Secret was not found")
)

func CreateSecret(clusterName, clusterEndpoint, credentials string) error {
	srv, err := secretservice.NewService()
	if err != nil {
		return err
	}

	session, err := srv.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return err
	}

	secret, err := session.NewSecret([]byte(credentials))
	if err != nil {
		return err
	}

	defer func() {
		_ = srv.LockItems([]dbus.ObjectPath{secretservice.DefaultCollection})
	}()
	if err := srv.Unlock([]dbus.ObjectPath{secretservice.DefaultCollection}); err != nil {
		return err
	}

	_, err = srv.CreateItem(
		secretservice.DefaultCollection,
		secretservice.NewSecretProperties(
			AccessGroup,
			map[string]string{
				"clusterName":     clusterName,
				"clusterEndpoint": clusterEndpoint,
			},
		),
		secret,
		secretservice.ReplaceBehaviorDoNotReplace,
	)

	return err
}

func DeleteSecret(clusterEndpoint string) error {
	srv, err := secretservice.NewService()
	if err != nil {
		return err
	}

	defer func() {
		_ = srv.LockItems([]dbus.ObjectPath{secretservice.DefaultCollection})
	}()
	if err := srv.Unlock([]dbus.ObjectPath{secretservice.DefaultCollection}); err != nil {
		return err
	}

	item, err := getSecretItem(clusterEndpoint)
	if err != nil {
		return err
	}

	return srv.DeleteItem(item)
}

func GetSecret(clusterEndpoint string) (string, string, error) {
	srv, err := secretservice.NewService()
	if err != nil {
		return "", "", err
	}

	session, err := srv.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return "", "", err
	}

	defer func() {
		_ = srv.LockItems([]dbus.ObjectPath{secretservice.DefaultCollection})
	}()
	if err := srv.Unlock([]dbus.ObjectPath{secretservice.DefaultCollection}); err != nil {
		return "", "", err
	}

	item, err := getSecretItem(clusterEndpoint)
	if err != nil {
		return "", "", err
	}

	attrs, err := srv.GetAttributes(item)
	if err != nil {
		return "", "", err
	}

	secret, err := srv.GetSecret(item, *session)
	if err != nil {
		return "", "", err
	}

	return attrs["clusterName"], string(secret), nil
}

func getSecretItem(clusterEndpoint string) (dbus.ObjectPath, error) {
	srv, err := secretservice.NewService()
	if err != nil {
		return "", err
	}

	defer func() {
		_ = srv.LockItems([]dbus.ObjectPath{secretservice.DefaultCollection})
	}()
	if err := srv.Unlock([]dbus.ObjectPath{secretservice.DefaultCollection}); err != nil {
		return "", err
	}

	items, err := srv.SearchCollection(
		secretservice.DefaultCollection,
		map[string]string{
			"clusterEndpoint": clusterEndpoint,
		},
	)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "", ErrorItemNotFound
	}
	if len(items) != 1 {
		return "", ErrorDuplicateItem
	}

	return items[0], nil
}
