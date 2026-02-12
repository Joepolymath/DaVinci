package chats

// Role constants for chat messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Config holds the configuration for the local LLM client.
type Config struct {
	Host  string // e.g. "http://localhost:11434" (Ollama default)
	Model string // e.g. "llama3:8b"
}

// IsValid returns true if the configuration has the minimum required fields.
func (c *Config) IsValid() bool {
	return c.Host != "" && c.Model != ""
}

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // The message content
}

// CompletionRequest is the payload sent to the local LLM for a chat completion.
type CompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  *Options  `json:"options,omitempty"`
}

// Options are optional model-level parameters.
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	MaxTokens   int     `json:"num_predict,omitempty"` // Ollama uses "num_predict"
	Stop        []string `json:"stop,omitempty"`
}

// CompletionResponse is the full (non-streaming) response from the local LLM.
type CompletionResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`

	// Usage statistics (populated when done=true)
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// StreamChunk is a single chunk received during streaming.
// During streaming, each line is a JSON object with partial content.
// The final chunk has Done=true and includes usage statistics.
type StreamChunk struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`

	// Only present in the final chunk (Done=true)
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

