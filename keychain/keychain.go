package keychain

import "github.com/keybase/go-keychain"

const (
	Service     = "Kubernetes"
	AccessGroup = "github.com/plumber-cd/kubectl-credentials-helper"
)

var (
	ErrorDuplicateItem = keychain.ErrorDuplicateItem
	ErrorItemNotFound  = keychain.ErrorItemNotFound
)
