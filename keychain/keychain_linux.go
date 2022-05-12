//go:build linux
// +build linux

package keychain

import (
	"errors"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/keybase/dbus"
	"github.com/keybase/go-keychain/secretservice"
)

var (
	ErrorDuplicateItem = errors.New("Secret already existed")
	ErrorItemNotFound  = errors.New("Secret was not found")
)

func LockUnlock() {

}

func CreateSecret(clusterName, clusterEndpoint, credentials string) error {
	if err := openItem(
		clusterEndpoint,
		func(
			_ *secretservice.SecretService,
			_ dbus.ObjectPath,
			_attrs secretservice.Attributes,
			_secret string,
		) error {
			return ErrorDuplicateItem
		},
	); err != nil && err != ErrorItemNotFound {
		return err
	}

	srv, err := secretservice.NewService()
	if err != nil {
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
			clusterEndpoint,
			map[string]string{
				"label":           AccessGroup,
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
	return openItem(
		clusterEndpoint,
		func(
			srv *secretservice.SecretService,
			item dbus.ObjectPath,
			_ secretservice.Attributes,
			_ string,
		) error {
			return srv.DeleteItem(item)
		},
	)
}

func GetSecret(clusterEndpoint string) (string, string, error) {
	var attrs secretservice.Attributes
	var secret string
	if err := openItem(
		clusterEndpoint,
		func(
			_ *secretservice.SecretService,
			_ dbus.ObjectPath,
			_attrs secretservice.Attributes,
			_secret string,
		) error {
			attrs = _attrs
			secret = _secret
			return nil
		},
	); err != nil {
		return "", "", err
	}

	return attrs["clusterName"], string(secret), nil
}

func openItem(
	clusterEndpoint string,
	callback func(
		*secretservice.SecretService,
		dbus.ObjectPath,
		secretservice.Attributes,
		string,
	) error,
) error {
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

	items, err := srv.SearchCollection(
		secretservice.DefaultCollection,
		map[string]string{
			"label":           AccessGroup,
			"clusterEndpoint": clusterEndpoint,
		},
	)
	if err != nil {
		return err
	}

	for _, item := range items {
		attrs, err := srv.GetAttributes(item)
		if err != nil {
			return err
		}
		if slices.Contains(maps.Keys(attrs), "label") && attrs["label"] == AccessGroup && attrs["clusterEndpoint"] == clusterEndpoint {
			session, err := srv.OpenSession(secretservice.AuthenticationDHAES)
			if err != nil {
				return err
			}

			secret, err := srv.GetSecret(item, *session)
			if err != nil {
				return err
			}
			return callback(srv, item, attrs, string(secret))
		}
	}

	return ErrorItemNotFound
}
