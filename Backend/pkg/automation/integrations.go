package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/elvisouma/salmar-ai/pkg/models"
)

// IntegrationManager handles external service integrations
type IntegrationManager struct {
	integrations map[string]Integration
	httpClient   *http.Client
	mu           sync.RWMutex
}

// Integration defines the interface for service integrations
type Integration interface {
	Name() string
	Execute(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)
	Validate(data map[string]interface{}) error
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager() *IntegrationManager {
	return &IntegrationManager{
		integrations: make(map[string]Integration),
		httpClient:   &http.Client{},
	}
}

// RegisterIntegration adds a new integration to the manager
func (m *IntegrationManager) RegisterIntegration(integration Integration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.integrations[integration.Name()] = integration
}

// ExecuteIntegration runs a specific integration with the provided data
func (m *IntegrationManager) ExecuteIntegration(ctx context.Context, name string, data map[string]interface{}) (map[string]interface{}, error) {
	m.mu.RLock()
	integration, exists := m.integrations[name]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("integration %s not found", name)
	}

	// Validate the input data
	if err := integration.Validate(data); err != nil {
		return nil, fmt.Errorf("invalid data for integration %s: %w", name, err)
	}

	// Execute the integration
	return integration.Execute(ctx, data)
}

// ProcessIntegrations handles all integrations specified in a request
func (m *IntegrationManager) ProcessIntegrations(ctx context.Context, req *models.Request) (*models.Response, error) {
	if req.Integrations == nil || len(req.Integrations) == 0 {
		return nil, nil // No integrations to process
	}

	response := &models.Response{
		Mode:   req.Mode,
		Intent: req.Intent,
	}

	// Process each integration
	for name, data := range req.Integrations {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			response.Errors = append(response.Errors, models.ErrorInfo{
				Code:    "integration_error",
				Message: fmt.Sprintf("Invalid data format for integration %s", name),
				Source:  "IntegrationManager",
			})
			continue
		}

		result, err := m.ExecuteIntegration(ctx, name, dataMap)
		if err != nil {
			response.Errors = append(response.Errors, models.ErrorInfo{
				Code:    "integration_error",
				Message: err.Error(),
				Source:  "IntegrationManager",
			})
			continue
		}

		// Add integration result to response
		if response.ModelInfo == nil {
			response.ModelInfo = make(map[string]interface{})
		}
		if response.ModelInfo["integrations"] == nil {
			response.ModelInfo["integrations"] = make(map[string]interface{})
		}

		integrations, _ := response.ModelInfo["integrations"].(map[string]interface{})
		integrations[name] = result
		response.ModelInfo["integrations"] = integrations
	}

	return response, nil
}

// WebhookIntegration implements a simple webhook integration
type WebhookIntegration struct {
	httpClient *http.Client
}

// NewWebhookIntegration creates a new webhook integration
func NewWebhookIntegration(client *http.Client) *WebhookIntegration {
	if client == nil {
		client = &http.Client{}
	}
	return &WebhookIntegration{httpClient: client}
}

// Name returns the integration name
func (w *WebhookIntegration) Name() string {
	return "webhook"
}

// Validate checks if the webhook data is valid
func (w *WebhookIntegration) Validate(data map[string]interface{}) error {
	url, ok := data["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	return nil
}

// Execute sends a webhook request
func (w *WebhookIntegration) Execute(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	url, _ := data["url"].(string)
	method, ok := data["method"].(string)
	if !ok || method == "" {
		method = "POST"
	}

	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		payload = make(map[string]interface{})
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	headers, ok := data["headers"].(map[string]interface{})
	if ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Send request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending webhook: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		// If response isn't JSON, create a simple status response
		responseData = map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		}
	}

	return responseData, nil
}

// EmailIntegration implements a simple email notification integration
type EmailIntegration struct{}

// NewEmailIntegration creates a new email integration
func NewEmailIntegration() *EmailIntegration {
	return &EmailIntegration{}
}

// Name returns the integration name
func (e *EmailIntegration) Name() string {
	return "email"
}

// Validate checks if the email data is valid
func (e *EmailIntegration) Validate(data map[string]interface{}) error {
	to, ok := data["to"].(string)
	if !ok || to == "" {
		return fmt.Errorf("email recipient is required")
	}

	subject, ok := data["subject"].(string)
	if !ok || subject == "" {
		return fmt.Errorf("email subject is required")
	}

	return nil
}

// Execute sends an email notification (mock implementation)
func (e *EmailIntegration) Execute(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// This is a mock implementation
	// In a real system, this would connect to an email service

	to, _ := data["to"].(string)
	subject, _ := data["subject"].(string)
	body, _ := data["body"].(string)

	// Log the email details (in a real system, this would send the email)
	fmt.Printf("Sending email to: %s\nSubject: %s\nBody: %s\n", to, subject, body)

	return map[string]interface{}{
		"status":  "sent",
		"to":      to,
		"subject": subject,
	}, nil
}