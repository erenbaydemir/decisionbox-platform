# Adding LLM Providers

> **Version**: 0.3.0

This guide shows how to add support for a new LLM service. You'll implement one Go interface method, register with metadata, and import in two files.

## Interface

```go
// libs/go-common/llm/provider.go
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Validate(ctx context.Context) error
}
```

`Validate` checks that credentials and configuration are valid without consuming tokens.
Use lightweight API calls (e.g., list models) when possible.
Called by the "Test Connection" button in the dashboard.

**ChatRequest:**

| Field | Type | Description |
|-------|------|-------------|
| `Model` | string | Model ID (may be overridden per-request) |
| `SystemPrompt` | string | System-level instruction |
| `Messages` | []Message | Conversation messages (`{Role, Content}`) |
| `MaxTokens` | int | Maximum output tokens |
| `Temperature` | float64 | Sampling temperature (0.0–1.0) |

**ChatResponse:**

| Field | Type | Description |
|-------|------|-------------|
| `Content` | string | Response text |
| `Model` | string | Model that generated this |
| `StopReason` | string | Why generation stopped |
| `Usage.InputTokens` | int | Input tokens consumed |
| `Usage.OutputTokens` | int | Output tokens generated |

## Step 1: Create the Package

```bash
mkdir -p providers/llm/myprovider
cd providers/llm/myprovider
go mod init github.com/decisionbox-io/decisionbox/providers/llm/myprovider
```

Add to `go.mod`:
```
require github.com/decisionbox-io/decisionbox/libs/go-common v0.0.0
replace github.com/decisionbox-io/decisionbox/libs/go-common => ../../../libs/go-common
```

## Step 2: Implement the Provider

```go
// providers/llm/myprovider/provider.go
package myprovider

import (
    "context"
    "fmt"
    "net/http"
    "strconv"
    "time"

    gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func init() {
    gollm.RegisterWithMeta("myprovider", func(cfg gollm.ProviderConfig) (gollm.Provider, error) {
        apiKey := cfg["api_key"]
        if apiKey == "" {
            return nil, fmt.Errorf("myprovider: api_key is required")
        }
        model := cfg["model"]
        if model == "" {
            model = "default-model"
        }
        timeoutSec, _ := strconv.Atoi(cfg["timeout_seconds"])
        if timeoutSec == 0 {
            timeoutSec = 300
        }

        return &MyProvider{
            apiKey:     apiKey,
            model:      model,
            httpClient: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
        }, nil
    }, gollm.ProviderMeta{
        Name:        "My LLM Provider",
        Description: "Description shown in the dashboard",
        ConfigFields: []gollm.ConfigField{
            {Key: "api_key", Label: "API Key", Required: true, Type: "string", Placeholder: "your-key-here"},
            {Key: "model", Label: "Model", Required: true, Type: "string", Default: "default-model"},
        },
        DefaultPricing: map[string]gollm.TokenPricing{
            "default-model": {InputPerMillion: 1.0, OutputPerMillion: 2.0},
        },
        MaxOutputTokens: map[string]int{
            "default-model": 4096,
            "_default":      4096,
        },
    })
}

type MyProvider struct {
    apiKey     string
    model      string
    httpClient *http.Client
}

func (p *MyProvider) Chat(ctx context.Context, req gollm.ChatRequest) (*gollm.ChatResponse, error) {
    model := req.Model
    if model == "" {
        model = p.model
    }
    maxTokens := req.MaxTokens
    if maxTokens == 0 {
        maxTokens = 4096
    }

    // TODO: Call your LLM API here
    // 1. Build request body from req.Messages, req.SystemPrompt, maxTokens
    // 2. Send HTTP request with p.apiKey
    // 3. Parse response
    // 4. Return ChatResponse

    return &gollm.ChatResponse{
        Content:    "response text",
        Model:      model,
        StopReason: "end_turn",
        Usage: gollm.Usage{
            InputTokens:  100,
            OutputTokens: 50,
        },
    }, nil
}
```

### Key Implementation Notes

- **Read `timeout_seconds` from config** — The agent passes this from the `LLM_TIMEOUT` env var
- **Support model override** — `req.Model` may differ from the provider default (per-request override)
- **Return accurate token counts** — Used for cost estimation and context tracking
- **Handle retries externally** — The agent's AI client handles retries. Your provider should not retry internally.
- **Set `MaxOutputTokens` accurately** — Check each model's API documentation for max output token limits. The agent calls `gollm.GetMaxOutputTokens(providerName, model)` to request the model's full output capacity during phases like recommendation generation. Use the `_default` key for a fallback when the exact model name is not listed.

### If your API speaks the OpenAI `/chat/completions` schema

Many providers expose an OpenAI-compatible chat completions API (OpenAI itself, Azure AI Foundry for OpenAI-family deployments, AWS Bedrock for Qwen, DeepSeek, Groq, Together, Fireworks, vLLM, LM Studio, etc.). For these, use the shared helper package at `libs/go-common/llm/openaicompat` instead of duplicating the request/response types. It gives you:

- `openaicompat.BuildRequestBody(req, model)` → marshals a `gollm.ChatRequest` into the canonical OpenAI body (system prompt prepended as `role="system"` message, `max_tokens` / `temperature` handled, empty model omitted)
- `openaicompat.ParseResponseBody(body)` → decodes the response and returns a `*gollm.ChatResponse`, surfacing structured `*openaicompat.APIError` errors via the standard `error` interface
- `openaicompat.ExtractAPIError(body)` → for transports that receive a non-200 status and want to render a structured error message from the body

Only the transport and authentication stay in your provider. Sketch:

```go
import "github.com/decisionbox-io/decisionbox/libs/go-common/llm/openaicompat"

func (p *MyProvider) Chat(ctx context.Context, req gollm.ChatRequest) (*gollm.ChatResponse, error) {
    model := req.Model
    if model == "" {
        model = p.model
    }

    body, err := openaicompat.BuildRequestBody(req, model)
    if err != nil {
        return nil, fmt.Errorf("myprovider: %w", err)
    }

    // Send `body` via your provider's transport (HTTP POST, AWS InvokeModel, etc.)
    // and authentication (Bearer header, signed request, api-key header, etc.).
    respBytes, status, err := p.send(ctx, body)
    if err != nil {
        return nil, fmt.Errorf("myprovider: request failed: %w", err)
    }
    if status != 200 {
        if apiErr := openaicompat.ExtractAPIError(respBytes); apiErr != nil {
            return nil, fmt.Errorf("myprovider: API error (%d): %s", status, apiErr.Error())
        }
        return nil, fmt.Errorf("myprovider: API error (%d): %s", status, string(respBytes))
    }

    resp, err := openaicompat.ParseResponseBody(respBytes)
    if err != nil {
        return nil, fmt.Errorf("myprovider: %w", err)
    }
    return resp, nil
}
```

If your API uses a different model selector (e.g., AWS Bedrock where the model ID sits in `InvokeModelInput.ModelId`), pass `""` as the second argument to `BuildRequestBody` so the `model` field is omitted from the JSON body. See `providers/llm/bedrock/qwen.go` for a full working example that combines `openaicompat` with a non-HTTP transport.

## Step 3: Register in Services

Add blank imports to both services:

```go
// services/agent/main.go
import _ "github.com/decisionbox-io/decisionbox/providers/llm/myprovider"

// services/api/main.go
import _ "github.com/decisionbox-io/decisionbox/providers/llm/myprovider"
```

Add `replace` directives to both `services/agent/go.mod` and `services/api/go.mod`:

```
require github.com/decisionbox-io/decisionbox/providers/llm/myprovider v0.0.0
replace github.com/decisionbox-io/decisionbox/providers/llm/myprovider => ../../providers/llm/myprovider
```

Update Dockerfiles to copy the go.mod (and go.sum if needed):

```dockerfile
# In services/agent/Dockerfile and services/api/Dockerfile
COPY providers/llm/myprovider/go.mod providers/llm/myprovider/
```

## Step 4: Write Tests

```go
// providers/llm/myprovider/provider_test.go
package myprovider

import (
    "testing"

    gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func TestRegistered(t *testing.T) {
    _, ok := gollm.GetProviderMeta("myprovider")
    if !ok {
        t.Fatal("myprovider not registered")
    }
}

func TestFactoryMissingKey(t *testing.T) {
    _, err := gollm.NewProvider("myprovider", gollm.ProviderConfig{})
    if err == nil {
        t.Fatal("should error without API key")
    }
}

func TestFactorySuccess(t *testing.T) {
    _, err := gollm.NewProvider("myprovider", gollm.ProviderConfig{
        "api_key": "test-key",
        "model":   "test-model",
    })
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

Add integration tests that skip without credentials:

```go
// providers/llm/myprovider/integration_test.go
//go:build integration

package myprovider

import (
    "context"
    "os"
    "testing"
    "time"

    gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
)

func TestIntegration_BasicChat(t *testing.T) {
    apiKey := os.Getenv("INTEGRATION_TEST_MYPROVIDER_API_KEY")
    if apiKey == "" {
        t.Skip("INTEGRATION_TEST_MYPROVIDER_API_KEY not set")
    }

    provider, _ := gollm.NewProvider("myprovider", gollm.ProviderConfig{
        "api_key": apiKey, "model": "default-model",
    })

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := provider.Chat(ctx, gollm.ChatRequest{
        Messages:  []gollm.Message{{Role: "user", Content: "Say hello in one word."}},
        MaxTokens: 10,
    })
    if err != nil {
        t.Fatalf("Chat error: %v", err)
    }
    if resp.Content == "" {
        t.Error("empty response")
    }
    t.Logf("Response: %q (tokens: in=%d out=%d)", resp.Content, resp.Usage.InputTokens, resp.Usage.OutputTokens)
}
```

## Step 5: Add to Makefile

Add your provider to the test targets:

```makefile
# In test-go target:
cd providers/llm/myprovider && go test ./...

# In test-llm target (integration):
cd providers/llm/myprovider && go test -tags=integration -count=1 -timeout=2m -v ./...
```

## Checklist

- [ ] `init()` registers with `RegisterWithMeta` (name, factory, metadata)
- [ ] `ConfigFields` includes all user-configurable fields
- [ ] `DefaultPricing` includes token pricing for common models
- [ ] `MaxOutputTokens` includes per-model max output token limits (with `_default` fallback)
- [ ] `timeout_seconds` read from config (not hardcoded)
- [ ] Model override supported (`req.Model` takes priority)
- [ ] Token usage returned accurately
- [ ] For OpenAI-compatible APIs: use `libs/go-common/llm/openaicompat` instead of copying the request/response types
- [ ] Imported in agent + API `main.go`
- [ ] `replace` directive in both go.mod files
- [ ] Dockerfile COPY line for go.mod
- [ ] Unit tests (registration, factory, config validation)
- [ ] Integration tests (skip without credentials)
- [ ] Added to Makefile test targets

## Next Steps

- [Providers Concept](../concepts/providers.md) — Plugin architecture overview
- [Configuring LLM Providers](configuring-llm.md) — How users set up LLM providers
