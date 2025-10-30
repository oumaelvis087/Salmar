package models

// RequestMode represents the primary mode of a request
type RequestMode string

const (
	ModeText  RequestMode = "text"
	ModeImage RequestMode = "image"
	ModeCode  RequestMode = "code"
	ModeMulti RequestMode = "multimodal"
)

// IntentType represents the classified intent of a request
type IntentType string

const (
	IntentGenerate    IntentType = "generate"
	IntentAnalyze     IntentType = "analyze"
	IntentExplain     IntentType = "explain"
	IntentSummarize   IntentType = "summarize"
	IntentTranslate   IntentType = "translate"
	IntentOptimize    IntentType = "optimize"
	IntentDebug       IntentType = "debug"
	IntentAutomate    IntentType = "automate"
	IntentConversation IntentType = "conversation"
)

// Request represents a multimodal request to the Salmar AI system
type Request struct {
	// Text input
	Text string `json:"text,omitempty"`
	
	// Image input
	ImageURL string `json:"image_url,omitempty"`
	ImageData string `json:"image_data,omitempty"` // Base64 encoded image
	
	// Code input
	Code     string `json:"code,omitempty"`
	Language string `json:"language,omitempty"`
	
	// Context and history
	ConversationID string        `json:"conversation_id,omitempty"`
	History        []ChatMessage `json:"history,omitempty"`
	UserID         string        `json:"user_id,omitempty"`
	
	// Processing options
	Mode    RequestMode `json:"mode,omitempty"`
	Intent  IntentType  `json:"intent,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
	
	// Automation and integration
	Schedule     string                 `json:"schedule,omitempty"`
	Integrations map[string]interface{} `json:"integrations,omitempty"`
	
	// User preferences
	Preferences UserPreferences `json:"preferences,omitempty"`
}

// Response represents a multimodal response from the Salmar AI system
type Response struct {
	// Mode and intent classification
	Mode   RequestMode `json:"mode,omitempty"`
	Intent IntentType  `json:"intent,omitempty"`
	
	// Text response
	TextResponse string `json:"text_response,omitempty"`
	
	// Image analysis and generation
	ImageAnalysis     map[string]interface{} `json:"image_analysis,omitempty"`
	GeneratedImageURL string                 `json:"generated_image_url,omitempty"`
	
	// Code generation and explanation
	GeneratedCode   string `json:"generated_code,omitempty"`
	CodeExplanation string `json:"code_explanation,omitempty"`
	
	// Reasoning and transparency
	ReasoningSteps []ReasoningStep `json:"reasoning_steps,omitempty"`
	Citations      []Citation      `json:"citations,omitempty"`
	
	// Metadata
	ProcessingTime float64                `json:"processing_time,omitempty"`
	ModelInfo      map[string]interface{} `json:"model_info,omitempty"`
	
	// Error handling
	Errors []ErrorInfo `json:"errors,omitempty"`
}

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`    // "user", "assistant", or "system"
	Content string `json:"content"` // The message content
	
	// Support for multimodal messages
	ImageURLs []string               `json:"image_urls,omitempty"`
	CodeBlocks []CodeBlock           `json:"code_blocks,omitempty"`
	Attachments []map[string]interface{} `json:"attachments,omitempty"`
	
	Timestamp int64  `json:"timestamp,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

// CodeBlock represents a block of code in a message
type CodeBlock struct {
	Code     string `json:"code"`
	Language string `json:"language,omitempty"`
}

// ReasoningStep represents a step in the AI's reasoning process
type ReasoningStep struct {
	Description string `json:"description"`
	Confidence  float64 `json:"confidence,omitempty"`
}

// Citation represents a source citation for information
type Citation struct {
	Text  string `json:"text"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"` // Which component generated the error
}

// UserPreferences stores user-specific preferences
type UserPreferences struct {
	TechnicalLevel string `json:"technical_level,omitempty"` // beginner, intermediate, advanced, expert
	ResponseStyle  string `json:"response_style,omitempty"`  // formal, conversational, educational, technical
	Language       string `json:"language,omitempty"`        // Preferred language for responses
}

// MemoryEntry represents a stored memory for context
type MemoryEntry struct {
	UserID        string    `json:"user_id"`
	ConversationID string   `json:"conversation_id"`
	Content       string    `json:"content"`
	Embedding     []float64 `json:"embedding,omitempty"`
	Timestamp     int64     `json:"timestamp"`
	Tags          []string  `json:"tags,omitempty"`
}