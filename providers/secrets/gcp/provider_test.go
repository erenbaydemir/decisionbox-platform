package gcp

import (
	"testing"

	"github.com/decisionbox-io/decisionbox/libs/go-common/secrets"
)

func TestGCPProvider_ImplementsInterface(t *testing.T) {
	var _ secrets.Provider = (*GCPProvider)(nil)
}

func TestSecretName(t *testing.T) {
	p := &GCPProvider{namespace: "decisionbox", gcpProject: "my-project"}

	name := p.secretName("proj-123", "llm-api-key")
	want := "decisionbox-proj-123-llm-api-key"
	if name != want {
		t.Errorf("secretName = %q, want %q", name, want)
	}
}

func TestSecretName_CustomNamespace(t *testing.T) {
	p := &GCPProvider{namespace: "myapp-prod", gcpProject: "my-project"}

	name := p.secretName("proj-456", "warehouse-creds")
	want := "myapp-prod-proj-456-warehouse-creds"
	if name != want {
		t.Errorf("secretName = %q, want %q", name, want)
	}
}

func TestSecretPath(t *testing.T) {
	p := &GCPProvider{namespace: "decisionbox", gcpProject: "my-gcp-project"}

	path := p.secretPath("proj-123", "llm-api-key")
	want := "projects/my-gcp-project/secrets/decisionbox-proj-123-llm-api-key"
	if path != want {
		t.Errorf("secretPath = %q, want %q", path, want)
	}
}

func TestVersionPath(t *testing.T) {
	p := &GCPProvider{namespace: "decisionbox", gcpProject: "my-gcp-project"}

	path := p.versionPath("proj-123", "llm-api-key")
	want := "projects/my-gcp-project/secrets/decisionbox-proj-123-llm-api-key/versions/latest"
	if path != want {
		t.Errorf("versionPath = %q, want %q", path, want)
	}
}

func TestNewGCPProviderWithClient_DefaultNamespace(t *testing.T) {
	p := NewGCPProviderWithClient(nil, "my-project", "")
	if p.namespace != "decisionbox" {
		t.Errorf("namespace = %q, want decisionbox", p.namespace)
	}
	if p.gcpProject != "my-project" {
		t.Errorf("gcpProject = %q, want my-project", p.gcpProject)
	}
}

func TestNewGCPProviderWithClient_CustomNamespace(t *testing.T) {
	p := NewGCPProviderWithClient(nil, "prod-project", "myapp-prod")
	if p.namespace != "myapp-prod" {
		t.Errorf("namespace = %q, want myapp-prod", p.namespace)
	}
}

func TestGCPProvider_Registered(t *testing.T) {
	registered := secrets.RegisteredProviders()
	found := false
	for _, name := range registered {
		if name == "gcp" {
			found = true
			break
		}
	}
	if !found {
		t.Error("gcp not registered in secrets provider registry")
	}
}

func TestGCPProvider_FactoryMissingProjectID(t *testing.T) {
	_, err := secrets.NewProvider(secrets.Config{
		Provider:  "gcp",
		Namespace: "test",
	})
	if err == nil {
		t.Error("should error without GCP project ID")
	}
}

func TestSecretName_NameFormat(t *testing.T) {
	// GCP secret names only allow: letters, numbers, hyphens, underscores
	// Verify our naming uses hyphens (not slashes like AWS)
	p := &GCPProvider{namespace: "decisionbox", gcpProject: "proj"}

	tests := []struct {
		projectID string
		key       string
		want      string
	}{
		{"proj-1", "llm-api-key", "decisionbox-proj-1-llm-api-key"},
		{"abc", "warehouse-credentials", "decisionbox-abc-warehouse-credentials"},
		{"p123", "key", "decisionbox-p123-key"},
	}
	for _, tt := range tests {
		got := p.secretName(tt.projectID, tt.key)
		if got != tt.want {
			t.Errorf("secretName(%q, %q) = %q, want %q", tt.projectID, tt.key, got, tt.want)
		}
	}
}

func TestSecretPath_IncludesGCPProject(t *testing.T) {
	p := &GCPProvider{namespace: "ns", gcpProject: "gcp-project-id"}

	path := p.secretPath("proj", "key")
	if path != "projects/gcp-project-id/secrets/ns-proj-key" {
		t.Errorf("path = %q", path)
	}
}

func TestVersionPath_IncludesLatest(t *testing.T) {
	p := &GCPProvider{namespace: "ns", gcpProject: "gcp-proj"}

	path := p.versionPath("proj", "key")
	if path != "projects/gcp-proj/secrets/ns-proj-key/versions/latest" {
		t.Errorf("path = %q", path)
	}
}
