// Package credentials stores provider API keys in the OS keyring.
package credentials

import "github.com/zalando/go-keyring"

// service namespaces our entries in the OS keyring.
const service = "shell-commander"

// Get returns the stored key for a provider. It returns keyring.ErrNotFound
// when nothing has been saved.
func Get(provider string) (string, error) {
	return keyring.Get(service, provider)
}

// Set stores the key for a provider, replacing any existing value.
func Set(provider, key string) error {
	return keyring.Set(service, provider, key)
}

// Delete removes the stored key for a provider.
func Delete(provider string) error {
	return keyring.Delete(service, provider)
}
