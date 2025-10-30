package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/elvisouma/salmar-ai/internal/config"
	"github.com/elvisouma/salmar-ai/pkg/automation"
	"github.com/elvisouma/salmar-ai/pkg/client"
	"github.com/elvisouma/salmar-ai/pkg/models"
)

// ContentFilter interface for ethical filtering
type ContentFilter interface {
	FilterResponse(response *models.Response) *models.Response
}

// ReasoningTransparency interface for enhancing reasoning transparency
type ReasoningTransparency interface {
	EnhanceTransparency(response *models.Response, request *models.Request) *models.Response
}

// TaskScheduler interface for automation
type TaskScheduler interface {
	AddTask(task *automation.ScheduledTask) error
	RemoveTask(taskID string) error
	Start(ctx context.Context) error
	Stop()
}

// IntegrationManager interface for external integrations
type IntegrationManager interface {
	ProcessIntegrations(ctx context.Context, req *models.Request) (*models.Response, error)
}

// Orchestrator coordinates requests between different services
type Orchestrator struct {
	config               *config.Config
	pythonClient         *client.PythonServiceClient
	memoryStore          MemoryStore
	contentFilter        ContentFilter
	reasoningTransparency ReasoningTransparency
	taskScheduler        TaskScheduler
	integrationManager   IntegrationManager
}

// MemoryStore interface for storing and retrieving context
type MemoryStore interface {
	StoreMemory(ctx context.Context, memory *models.MemoryEntry) error
	RetrieveMemories(ctx context.Context, userID, conversationID string, limit int) ([]models.MemoryEntry, error)
	SearchSimilarMemories(ctx context.Context, query string, limit int) ([]models.MemoryEntry, error)
}

// New creates a new orchestrator instance
func New(cfg *config.Config, pythonClient *client.PythonServiceClient, memoryStore MemoryStore) *Orchestrator {
	return &Orchestrator{
		config:       cfg,
		pythonClient: pythonClient,
		memoryStore:  memoryStore,
	}
}

// SetContentFilter sets the content filter component
func (o *Orchestrator) SetContentFilter(filter ContentFilter) {
	o.contentFilter = filter
}

// SetReasoningTransparency sets the reasoning transparency component
func (o *Orchestrator) SetReasoningTransparency(transparency ReasoningTransparency) {
	o.reasoningTransparency = transparency
}

// SetTaskScheduler sets the task scheduler component
func (o *Orchestrator) SetTaskScheduler(scheduler TaskScheduler) {
	o.taskScheduler = scheduler
}

// SetIntegrationManager sets the integration manager component
func (o *Orchestrator) SetIntegrationManager(manager IntegrationManager) {
	o.integrationManager = manager
}

// ClassifyIntent determines the primary intent of the request
func (o *Orchestrator) ClassifyIntent(ctx context.Context, req *models.Request) (models.IntentType, error) {
	// If intent is already specified, use it
	if req.Intent != "" {
		return req.Intent, nil
	}

	// Call NLP service to classify intent
	intentReq := &models.Request{
		Text: req.Text,
		Mode: models.ModeText,
	}
	
	url := fmt.Sprintf("http://%s:%s/classify_intent", o.config.PythonServicesHost, o.config.NLPServicePort)
	resp, err := o.callService(ctx, url, intentReq)
	if err != nil {
		// Default to conversation if classification fails
		return models.IntentConversation, fmt.Errorf("intent classification error: %w", err)
	}
	
	return resp.Intent, nil
}

// DetermineMode identifies the primary mode of the request
func (o *Orchestrator) DetermineMode(req *models.Request) models.RequestMode {
	// If mode is already specified, use it
	if req.Mode != "" {
		return req.Mode
	}
	
	// Determine mode based on input fields
	hasText := req.Text != ""
	hasImage := req.ImageURL != "" || req.ImageData != ""
	hasCode := req.Code != ""
	
	if hasText && !hasImage && !hasCode {
		return models.ModeText
	} else if hasImage && !hasText && !hasCode {
		return models.ModeImage
	} else if hasCode && !hasText && !hasImage {
		return models.ModeCode
	} else if hasText || hasImage || hasCode {
		return models.ModeMulti
	}
	
	// Default to text mode if no inputs are provided
	return models.ModeText
}

// ProcessRequest handles a multimodal request by routing to appropriate services
func (o *Orchestrator) ProcessRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(o.config.RequestTimeout)*time.Second)
	defer cancel()

	// Start timing for processing
	startTime := time.Now()
	
	// Determine request mode and intent
	req.Mode = o.DetermineMode(req)
	
	// Classify intent if not specified
	if req.Intent == "" {
		intent, err := o.ClassifyIntent(ctx, req)
		if err != nil {
			// Log the error but continue with default intent
			fmt.Printf("Intent classification error: %v\n", err)
		}
		req.Intent = intent
	}
	
	// Check for integration requests first
	if o.integrationManager != nil && req.Intent == models.IntentAutomate {
		integrationResp, err := o.integrationManager.ProcessIntegrations(ctx, req)
		if err == nil && integrationResp != nil {
			// Apply ethical filtering and reasoning transparency
			return o.enhanceAndFilterResponse(integrationResp, req, time.Since(startTime)), nil
		}
		// If integration processing fails, continue with normal processing
	}
	
	// Retrieve relevant context from memory if conversation ID is provided
	if req.ConversationID != "" && req.UserID != "" {
		memories, err := o.memoryStore.RetrieveMemories(ctx, req.UserID, req.ConversationID, 10)
		if err == nil && len(memories) > 0 {
			// Add context to request options if not already present
			if req.Options == nil {
				req.Options = make(map[string]interface{})
			}
			req.Options["context_memories"] = memories
		}
	}

	// Initialize response with request metadata
	response := models.Response{
		Mode:   req.Mode,
		Intent: req.Intent,
	}

	// Process based on mode and intent
	var processingErr error
	
	switch req.Mode {
	case models.ModeText:
		nlpResp, err := o.pythonClient.CallNLPService(ctx, req)
		if err != nil {
			processingErr = fmt.Errorf("NLP service error: %w", err)
		} else {
			response.TextResponse = nlpResp.TextResponse
			response.ReasoningSteps = nlpResp.ReasoningSteps
			response.Citations = nlpResp.Citations
		}
		
	case models.ModeImage:
		imgResp, err := o.pythonClient.CallImageService(ctx, req)
		if err != nil {
			processingErr = fmt.Errorf("image service error: %w", err)
		} else {
			response.ImageAnalysis = imgResp.ImageAnalysis
			response.GeneratedImageURL = imgResp.GeneratedImageURL
			response.TextResponse = imgResp.TextResponse
		}
		
	case models.ModeCode:
		codeResp, err := o.pythonClient.CallCodeService(ctx, req)
		if err != nil {
			processingErr = fmt.Errorf("code service error: %w", err)
		} else {
			response.GeneratedCode = codeResp.GeneratedCode
			response.CodeExplanation = codeResp.CodeExplanation
			response.ReasoningSteps = codeResp.ReasoningSteps
		}
		
	case models.ModeMulti:
		// For multimodal requests, use the PythonServiceClient to call multiple services
		multiResp, err := o.pythonClient.CallMultipleServices(ctx, req)
		if err != nil {
			processingErr = fmt.Errorf("multimodal processing error: %w", err)
		} else {
			// Copy all fields from multiResp to response
			response = *multiResp
		}
		
		// Collect results and errors
		var errors []error
		var responses []*models.Response
		
		serviceCount := 0
		if req.Text != "" {
			serviceCount++
		}
		if req.ImageURL != "" || req.ImageData != "" {
			serviceCount++
		}
		if req.Code != "" {
			serviceCount++
		}
		
		// Define channels for collecting responses and errors
		errChan := make(chan error, serviceCount)
		respChan := make(chan *models.Response, serviceCount)
		
		for i := 0; i < serviceCount; i++ {
			err := <-errChan
			resp := <-respChan
			
			if err != nil {
				errors = append(errors, err)
			}
			if resp != nil {
				responses = append(responses, resp)
			}
		}
		
		// Combine responses
		for _, resp := range responses {
			if resp.TextResponse != "" {
				response.TextResponse += resp.TextResponse + "\n\n"
			}
			if resp.GeneratedImageURL != "" {
				response.GeneratedImageURL = resp.GeneratedImageURL
			}
			if resp.ImageAnalysis != nil {
				response.ImageAnalysis = resp.ImageAnalysis
			}
			if resp.GeneratedCode != "" {
				response.GeneratedCode = resp.GeneratedCode
			}
			if resp.CodeExplanation != "" {
				response.CodeExplanation = resp.CodeExplanation
			}
			if len(resp.ReasoningSteps) > 0 {
				response.ReasoningSteps = append(response.ReasoningSteps, resp.ReasoningSteps...)
			}
			if len(resp.Citations) > 0 {
				response.Citations = append(response.Citations, resp.Citations...)
			}
		}
		
		// If all services failed, return an error
		if len(errors) == serviceCount {
			processingErr = fmt.Errorf("all services failed: %v", errors)
		} else if len(errors) > 0 {
			// Add errors to response but don't fail completely
			for _, err := range errors {
				response.Errors = append(response.Errors, models.ErrorInfo{
					Code:    "SERVICE_ERROR",
					Message: err.Error(),
				})
			}
		}
	}
	
	// Apply ethical filtering and reasoning transparency
	enhancedResponse := o.enhanceAndFilterResponse(&response, req, time.Since(startTime))
	
	// Store interaction in memory if conversation ID is provided
	if req.ConversationID != "" && req.UserID != "" && enhancedResponse.TextResponse != "" {
		memory := &models.MemoryEntry{
			UserID:        req.UserID,
			ConversationID: req.ConversationID,
			Content:       enhancedResponse.TextResponse,
			Timestamp:     time.Now().Unix(),
		}
		
		// Async store to not block response
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = o.memoryStore.StoreMemory(ctx, memory)
		}()
	}
	
	// Return error if processing failed completely
	if processingErr != nil && len(enhancedResponse.Errors) == 0 {
		return nil, processingErr
	}
	
	return enhancedResponse, nil
}

// callNLPService sends a request to the NLP service
func (o *Orchestrator) callNLPService(ctx context.Context, req *models.Request) (*models.Response, error) {
	url := fmt.Sprintf("http://%s:%s/process", o.config.PythonServicesHost, o.config.NLPServicePort)
	return o.callService(ctx, url, req)
}

// callImageService sends a request to the image processing service
func (o *Orchestrator) callImageService(ctx context.Context, req *models.Request) (*models.Response, error) {
	url := fmt.Sprintf("http://%s:%s/process", o.config.PythonServicesHost, o.config.ImageServicePort)
	return o.callService(ctx, url, req)
}

// callCodeService sends a request to the code generation/analysis service
func (o *Orchestrator) callCodeService(ctx context.Context, req *models.Request) (*models.Response, error) {
	url := fmt.Sprintf("http://%s:%s/process", o.config.PythonServicesHost, o.config.CodeServicePort)
	return o.callService(ctx, url, req)
}

// callService is a helper function to make HTTP requests to services
func (o *Orchestrator) callService(ctx context.Context, url string, req *models.Request) (*models.Response, error) {
	// Convert request to JSON
	reqBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("service returned non-200 status code")
	}

	// Parse response
	var response models.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// enhanceAndFilterResponse applies ethical filtering and reasoning transparency to the response
func (o *Orchestrator) enhanceAndFilterResponse(response *models.Response, req *models.Request, processingTime time.Duration) *models.Response {
	// Add processing time to response
	response.ProcessingTime = processingTime.Seconds()
	
	// Handle Gemini API specific model info if present
	if response.ModelInfo != nil {
		// Ensure we have a structured model_info section
		if _, ok := response.ModelInfo["gemini"]; !ok && 
		   (response.ModelInfo["model"] == "gemini-1.5-pro" || 
		    response.ModelInfo["model"] == "gemini-1.5-pro-vision") {
			
			// Restructure model info for consistency
			geminiInfo := make(map[string]interface{})
			for k, v := range response.ModelInfo {
				geminiInfo[k] = v
			}
			
			response.ModelInfo = map[string]interface{}{
				"gemini": geminiInfo,
				"provider": "google",
				"api_version": "v1",
			}
		}
		
		// Add processing metadata
		response.ModelInfo["processing_details"] = map[string]interface{}{
			"orchestrator": "salmar-go",
			"timestamp": time.Now().Format(time.RFC3339),
		}
	}
	
	// Apply ethical filtering if available
	if o.contentFilter != nil {
		response = o.contentFilter.FilterResponse(response)
	}
	
	// Apply reasoning transparency if available
	if o.reasoningTransparency != nil {
		response = o.reasoningTransparency.EnhanceTransparency(response, req)
	}
	
	return response
}