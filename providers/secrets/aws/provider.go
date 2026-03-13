// Package aws provides a secrets.Provider backed by AWS Secrets Manager.
// Status: STUB — registers the provider so it appears in the registry.
package aws

import (
	"context"
	"fmt"

	"github.com/decisionbox-io/decisionbox/libs/go-common/secrets"
)

func init() {
	secrets.Register("aws", func(cfg secrets.Config) (secrets.Provider, error) {
		return &AWSProvider{region: cfg.AWSRegion, namespace: cfg.Namespace}, nil
	}, secrets.ProviderMeta{
		Name:        "AWS Secrets Manager",
		Description: "Production secrets via AWS Secrets Manager — coming soon",
	})
}

type AWSProvider struct {
	region    string
	namespace string
}

func (p *AWSProvider) Get(ctx context.Context, projectID, key string) (string, error) {
	return "", fmt.Errorf("aws secret provider not yet implemented — configure SECRET_PROVIDER=mongodb for local dev (region=%s)", p.region)
}

func (p *AWSProvider) Set(ctx context.Context, projectID, key, value string) error {
	return fmt.Errorf("aws secret provider not yet implemented — configure SECRET_PROVIDER=mongodb for local dev")
}

func (p *AWSProvider) List(ctx context.Context, projectID string) ([]secrets.SecretEntry, error) {
	// Return empty with warning so UI shows a clear message
	return []secrets.SecretEntry{
		{Key: "(none)", Masked: "***", Warning: "AWS Secrets Manager provider not yet implemented. Use SECRET_PROVIDER=mongodb for now."},
	}, nil
}
