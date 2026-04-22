package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
	"github.com/decisionbox-io/decisionbox/libs/go-common/llm/openaicompat"
)

// qwenChat sends a request to a Qwen model on Bedrock using InvokeModel.
//
// Qwen on Bedrock accepts the standard OpenAI chat completions request and
// response body, so the body encoding and decoding is shared with the OpenAI
// and Azure AI Foundry providers via libs/go-common/llm/openaicompat. Only the
// transport (Bedrock InvokeModel vs. HTTP POST) differs between providers.
//
// The model identifier is carried by Bedrock's ModelId envelope field, so it is
// omitted from the JSON body to avoid unnecessary payload.
func (p *BedrockProvider) qwenChat(ctx context.Context, req gollm.ChatRequest) (*gollm.ChatResponse, error) {
	reqBody, err := openaicompat.BuildRequestBody(req, "")
	if err != nil {
		return nil, fmt.Errorf("bedrock/qwen: failed to marshal request: %w", err)
	}

	output, err := p.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        reqBody,
	})
	if err != nil {
		return nil, fmt.Errorf("bedrock/qwen: InvokeModel failed: %w", err)
	}

	resp, err := openaicompat.ParseResponseBody(output.Body)
	if err != nil {
		return nil, fmt.Errorf("bedrock/qwen: %w", err)
	}

	// Qwen responses may omit the model field; fall back to the requested model
	// so downstream telemetry can still attribute the call.
	if resp.Model == "" {
		resp.Model = req.Model
	}
	return resp, nil
}
