// Package azure provides a secrets.Provider backed by Azure Key Vault.
// Status: STUB — registers the provider so it appears in the registry.
package azure

import (
	"context"
	"fmt"

	"github.com/decisionbox-io/decisionbox/libs/go-common/secrets"
)

func init() {
	secrets.Register("azure", func(cfg secrets.Config) (secrets.Provider, error) {
		return &AzureProvider{vaultURL: cfg.AzureVaultURL, namespace: cfg.Namespace}, nil
	}, secrets.ProviderMeta{
		Name:        "Azure Key Vault",
		Description: "Production secrets via Azure Key Vault — coming soon",
	})
}

type AzureProvider struct {
	vaultURL  string
	namespace string
}

func (p *AzureProvider) Get(ctx context.Context, projectID, key string) (string, error) {
	return "", fmt.Errorf("azure secret provider not yet implemented — configure SECRET_PROVIDER=mongodb for local dev (vault=%s)", p.vaultURL)
}

func (p *AzureProvider) Set(ctx context.Context, projectID, key, value string) error {
	return fmt.Errorf("azure secret provider not yet implemented — configure SECRET_PROVIDER=mongodb for local dev")
}

func (p *AzureProvider) List(ctx context.Context, projectID string) ([]secrets.SecretEntry, error) {
	return []secrets.SecretEntry{
		{Key: "(none)", Masked: "***", Warning: "Azure Key Vault provider not yet implemented. Use SECRET_PROVIDER=mongodb for now."},
	}, nil
}
