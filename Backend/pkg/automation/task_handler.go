package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/elvisouma/salmar-ai/pkg/client"
	"github.com/elvisouma/salmar-ai/pkg/models"
)

// AITaskHandler implements TaskHandler for AI processing tasks
type AITaskHandler struct {
	client *client.PythonServiceClient
}

// NewAITaskHandler creates a new AI task handler
func NewAITaskHandler(client *client.PythonServiceClient) *AITaskHandler {
	return &AITaskHandler{
		client: client,
	}
}

// Execute runs the AI task
func (h *AITaskHandler) Execute(ctx context.Context, task *ScheduledTask) error {
	if task.Request == nil {
		return fmt.Errorf("task request cannot be nil")
	}

	// Add execution metadata to the request
	if task.Request.Options == nil {
		task.Request.Options = make(map[string]interface{})
	}
	task.Request.Options["scheduled_execution"] = true
	task.Request.Options["task_id"] = task.ID
	task.Request.Options["execution_time"] = time.Now().Format(time.RFC3339)

	// Process the request using the Python client
	var response *models.Response
	var err error

	// Call the appropriate service based on the request mode
	switch task.Request.Mode {
	case models.ModeMulti:
		response, err = h.client.CallMultipleServices(ctx, task.Request)
	case models.ModeText:
		response, err = h.client.CallNLPService(ctx, task.Request)
	case models.ModeImage:
		response, err = h.client.CallImageService(ctx, task.Request)
	case models.ModeCode:
		response, err = h.client.CallCodeService(ctx, task.Request)
	default:
		// Default to NLP service
		response, err = h.client.CallNLPService(ctx, task.Request)
	}

	if err != nil {
		return fmt.Errorf("error executing AI task: %w", err)
	}

	// Process any integrations in the request
	if task.Request.Integrations != nil && len(task.Request.Integrations) > 0 {
		// In a real implementation, this would handle the integrations
		// For now, we'll just log that integrations would be processed
		fmt.Printf("Processing %d integrations for task %s\n", len(task.Request.Integrations), task.ID)
	}

	// Log successful execution
	fmt.Printf("Successfully executed task %s (%s)\n", task.ID, task.Name)
	if response != nil {
		fmt.Printf("Response mode: %s, intent: %s\n", response.Mode, response.Intent)
	}

	return nil
}

// IntegrationTaskHandler implements TaskHandler for integration tasks
type IntegrationTaskHandler struct {
	integrationManager *IntegrationManager
}

// NewIntegrationTaskHandler creates a new integration task handler
func NewIntegrationTaskHandler(manager *IntegrationManager) *IntegrationTaskHandler {
	return &IntegrationTaskHandler{
		integrationManager: manager,
	}
}

// Execute runs the integration task
func (h *IntegrationTaskHandler) Execute(ctx context.Context, task *ScheduledTask) error {
	if task.Request == nil {
		return fmt.Errorf("task request cannot be nil")
	}

	if task.Request.Integrations == nil || len(task.Request.Integrations) == 0 {
		return fmt.Errorf("no integrations specified in task")
	}

	// Process the integrations
	response, err := h.integrationManager.ProcessIntegrations(ctx, task.Request)
	if err != nil {
		return fmt.Errorf("error processing integrations: %w", err)
	}

	// Check for errors in the response
	if response != nil && len(response.Errors) > 0 {
		return fmt.Errorf("integration errors occurred: %d errors", len(response.Errors))
	}

	// Log successful execution
	fmt.Printf("Successfully executed integration task %s (%s)\n", task.ID, task.Name)
	return nil
}