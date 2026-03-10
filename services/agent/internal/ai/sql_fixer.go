package ai

import (
	"context"
	"fmt"
	"strings"

	gollm "github.com/decisionbox-io/decisionbox/libs/go-common/llm"
	logger "github.com/decisionbox-io/decisionbox/services/agent/internal/log"
)

// SQLFixer uses LLM to fix SQL query errors.
type SQLFixer struct {
	client       *Client
	sqlFixPrompt string
	dataset      string
	filter       string
	schemaCtx    string
}

// SQLFixerOptions configures the SQL fixer.
type SQLFixerOptions struct {
	Client       *Client
	SQLFixPrompt string // from warehouse.Provider.SQLFixPrompt()
	Dataset      string
	Filter       string
}

// NewSQLFixer creates a new SQL fixer.
func NewSQLFixer(opts SQLFixerOptions) *SQLFixer {
	return &SQLFixer{
		client:       opts.Client,
		sqlFixPrompt: opts.SQLFixPrompt,
		dataset:      opts.Dataset,
		filter:       opts.Filter,
	}
}

// FixSQL attempts to fix a SQL query based on the error message.
func (f *SQLFixer) FixSQL(ctx context.Context, query string, errorMsg string, attempt int) (string, error) {
	logger.WithFields(logger.Fields{
		"attempt": attempt,
		"error":   errorMsg,
	}).Info("Attempting to fix SQL query")

	systemPrompt := f.sqlFixPrompt
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{DATASET}}", f.dataset)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{FILTER}}", f.filter)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{SCHEMA_INFO}}", f.schemaCtx)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{ORIGINAL_SQL}}", query)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{ERROR_MESSAGE}}", errorMsg)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{CONVERSATION_HISTORY}}", "")

	userMessage := fmt.Sprintf("Fix this SQL query (attempt %d). Return ONLY the corrected SQL.\n\nQuery:\n```sql\n%s\n```\n\nError:\n```\n%s\n```", attempt+1, query, errorMsg)

	conversation := NewConversation(ConversationOptions{
		SystemPrompt: systemPrompt,
		MaxMessages:  10,
	})
	conversation.AddUserMessage(userMessage)

	response, err := f.client.CreateMessage(ctx, conversation.GetMessages(), conversation.GetSystemPrompt(), 4000)
	if err != nil {
		return "", fmt.Errorf("failed to get SQL fix: %w", err)
	}

	fixedSQL, err := extractFixedSQL(response)
	if err != nil {
		return "", fmt.Errorf("failed to extract fixed SQL: %w", err)
	}

	logger.WithField("attempt", attempt).Info("SQL query fixed")

	return fixedSQL, nil
}

// SetSchemaContext updates the schema context.
func (f *SQLFixer) SetSchemaContext(schemaJSON string) {
	f.schemaCtx = schemaJSON
}

func extractFixedSQL(response *gollm.ChatResponse) (string, error) {
	if response == nil || response.Content == "" {
		return "", fmt.Errorf("empty response")
	}

	text := response.Content

	if strings.Contains(text, "```sql") {
		if sql := extractCodeBlock(text, "sql"); sql != "" {
			return strings.TrimSpace(sql), nil
		}
	}

	if strings.Contains(text, "```") {
		if sql := extractCodeBlock(text, ""); sql != "" {
			return strings.TrimSpace(sql), nil
		}
	}

	sql := strings.TrimSpace(text)
	if !strings.Contains(strings.ToUpper(sql), "SELECT") {
		return "", fmt.Errorf("response does not appear to be SQL")
	}

	return sql, nil
}

func extractCodeBlock(text string, language string) string {
	marker := "```"
	if language != "" {
		marker = "```" + language
	}

	startIdx := strings.Index(text, marker)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(marker)
	for startIdx < len(text) && (text[startIdx] == '\n' || text[startIdx] == '\r') {
		startIdx++
	}

	endIdx := strings.Index(text[startIdx:], "```")
	if endIdx == -1 {
		return ""
	}

	return text[startIdx : startIdx+endIdx]
}
