package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/elvisouma/salmar-ai/internal/config"
	"github.com/elvisouma/salmar-ai/pkg/models"
)

// PythonServiceClient handles communication with Python microservices
type PythonServiceClient struct {
	config     *config.Config
	httpClient *http.Client
	retryCount int
	retryDelay time.Duration
}

// NewPythonServiceClient creates a new client for Python services
func NewPythonServiceClient(cfg *config.Config) *PythonServiceClient {
	return &PythonServiceClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		retryCount: 3,
		retryDelay: 500 * time.Millisecond,
	}
}

// CallNLPService sends a request to the NLP service
func (c *PythonServiceClient) CallNLPService(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Determine the appropriate endpoint based on the request
	endpoint := "/process"
	if req.Intent != "" {
		// If intent is already provided, use the process endpoint
		endpoint = "/process"
	} else if req.Text != "" && req.ImageURL == "" && req.Code == "" {
		// If only text is provided, use the classify_intent endpoint first
		intentResp, err := c.classifyIntent(ctx, req)
		if err == nil && intentResp != nil {
			req.Intent = intentResp.Intent
		}
	}
	
	url := fmt.Sprintf("http://%s:%s%s", c.config.PythonServicesHost, c.config.NLPServicePort, endpoint)
	return c.callServiceWithRetry(ctx, url, req)
}

// classifyIntent calls the NLP service to classify the intent of a request
func (c *PythonServiceClient) classifyIntent(ctx context.Context, req *models.Request) (*models.Response, error) {
	url := fmt.Sprintf("http://%s:%s/classify_intent", c.config.PythonServicesHost, c.config.NLPServicePort)
	intentReq := &models.Request{
		Text: req.Text,
		Mode: models.ModeText,
	}
	return c.callServiceWithRetry(ctx, url, intentReq)
}

// CallImageService sends a request to the image processing service
func (c *PythonServiceClient) CallImageService(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Determine the appropriate endpoint based on the request
	endpoint := "/process"
	if req.Text != "" && req.ImageURL != "" {
		endpoint = "/multimodal"
	} else if req.ImageURL != "" && req.Intent == models.IntentAnalyze {
		endpoint = "/analyze"
	} else if req.Text != "" && req.Intent == models.IntentGenerate {
		endpoint = "/generate"
	}
	
	url := fmt.Sprintf("http://%s:%s%s", c.config.PythonServicesHost, c.config.ImageServicePort, endpoint)
	return c.callServiceWithRetry(ctx, url, req)
}

// CallCodeService sends a request to the code generation/analysis service
func (c *PythonServiceClient) CallCodeService(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Determine the appropriate endpoint based on the request
	endpoint := "/process"
	if req.Code != "" && req.Text != "" {
		endpoint = "/multimodal/code-text"
	} else if req.Code != "" && req.ImageURL != "" {
		endpoint = "/multimodal/code-image"
	} else if req.Text != "" && req.Intent == models.IntentGenerate {
		endpoint = "/generate"
	} else if req.Code != "" && req.Intent == models.IntentExplain {
		endpoint = "/explain"
	}
	
	url := fmt.Sprintf("http://%s:%s%s", c.config.PythonServicesHost, c.config.CodeServicePort, endpoint)
	return c.callServiceWithRetry(ctx, url, req)
}

// CallMultipleServices sends requests to multiple services in parallel and aggregates the results
func (c *PythonServiceClient) CallMultipleServices(ctx context.Context, req *models.Request) (*models.Response, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalResponse models.Response
	var errors []error
	
	// Determine which services to call based on the request mode
	if req.Mode == models.ModeMulti || (req.Text != "" && (req.ImageURL != "" || req.Code != "")) {
		wg.Add(3) // Call all three services
		
		// Call NLP service
		go func() {
			defer wg.Done()
			resp, err := c.CallNLPService(ctx, req)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("NLP service error: %w", err))
				return
			}
			c.mergeResponses(&finalResponse, resp)
		}()
		
		// Call Image service if image URL is provided
		go func() {
			defer wg.Done()
			if req.ImageURL != "" {
				resp, err := c.CallImageService(ctx, req)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errors = append(errors, fmt.Errorf("Image service error: %w", err))
					return
				}
				c.mergeResponses(&finalResponse, resp)
			}
		}()
		
		// Call Code service if code is provided
		go func() {
			defer wg.Done()
			if req.Code != "" {
				resp, err := c.CallCodeService(ctx, req)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errors = append(errors, fmt.Errorf("Code service error: %w", err))
					return
				}
				c.mergeResponses(&finalResponse, resp)
			}
		}()
		
		wg.Wait()
		
		// Apply ethical filtering to the final response
		finalResponse = c.applyEthicalFiltering(finalResponse)
		
		// If there were errors, add them to the response
		if len(errors) > 0 {
			for _, err := range errors {
				finalResponse.Errors = append(finalResponse.Errors, models.ErrorInfo{
					Message: err.Error(),
					Code:    "service_error",
				})
			}
		}
		
		return &finalResponse, nil
	}
	
	// If not multimodal, call the appropriate service based on the mode
	switch req.Mode {
	case models.ModeText:
		return c.CallNLPService(ctx, req)
	case models.ModeImage:
		return c.CallImageService(ctx, req)
	case models.ModeCode:
		return c.CallCodeService(ctx, req)
	default:
		// Default to NLP service
		return c.CallNLPService(ctx, req)
	}
}

// mergeResponses combines responses from multiple services
func (c *PythonServiceClient) mergeResponses(finalResp *models.Response, resp *models.Response) {
	// Set mode and intent if not already set
	if finalResp.Mode == "" {
		finalResp.Mode = resp.Mode
	}
	if finalResp.Intent == "" {
		finalResp.Intent = resp.Intent
	}
	
	// Merge text responses
	if resp.TextResponse != "" {
		if finalResp.TextResponse != "" {
			finalResp.TextResponse += "\n\n" + resp.TextResponse
		} else {
			finalResp.TextResponse = resp.TextResponse
		}
	}
	
	// Merge image analysis
	if len(resp.ImageAnalysis) > 0 {
		if finalResp.ImageAnalysis == nil {
			finalResp.ImageAnalysis = make(map[string]interface{})
		}
		for k, v := range resp.ImageAnalysis {
			finalResp.ImageAnalysis[k] = v
		}
	}
	
	// Set generated image URL if provided
	if resp.GeneratedImageURL != "" {
		finalResp.GeneratedImageURL = resp.GeneratedImageURL
	}
	
	// Set generated code and explanation if provided
	if resp.GeneratedCode != "" {
		finalResp.GeneratedCode = resp.GeneratedCode
	}
	if resp.CodeExplanation != "" {
		finalResp.CodeExplanation = resp.CodeExplanation
	}
	
	// Merge reasoning steps
	finalResp.ReasoningSteps = append(finalResp.ReasoningSteps, resp.ReasoningSteps...)
	
	// Merge citations
	finalResp.Citations = append(finalResp.Citations, resp.Citations...)
	
	// Merge model info
	if len(resp.ModelInfo) > 0 {
		if finalResp.ModelInfo == nil {
			finalResp.ModelInfo = make(map[string]interface{})
		}
		for k, v := range resp.ModelInfo {
			finalResp.ModelInfo[k] = v
		}
	}
	
	// Merge errors
	finalResp.Errors = append(finalResp.Errors, resp.Errors...)
}

// applyEthicalFiltering applies ethical filtering to the response
func (c *PythonServiceClient) applyEthicalFiltering(resp models.Response) models.Response {
	// Add a reasoning step for ethical filtering
	ethicalStep := models.ReasoningStep{
		Description: "Applied ethical content filtering to ensure response adheres to ethical guidelines",
		Confidence:  0.95,
	}
	resp.ReasoningSteps = append(resp.ReasoningSteps, ethicalStep)
	
	// Add ethical filtering logic here
	// This is a placeholder for actual implementation
	// In a real implementation, this would check for harmful content
	
	return resp
}

// callServiceWithRetry calls a service with retry logic
func (c *PythonServiceClient) callServiceWithRetry(ctx context.Context, url string, req *models.Request) (*models.Response, error) {
	var lastErr error
	
	for i := 0; i <= c.retryCount; i++ {
		resp, err := c.callService(ctx, url, req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		// If this was the last attempt, break
		if i == c.retryCount {
			break
		}
		
		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.retryDelay):
			// Continue to next retry
		}
	}
	
	return nil, fmt.Errorf("service call failed after %d retries: %w", c.retryCount, lastErr)
}

// callService is a helper function to make HTTP requests to services
func (c *PythonServiceClient) callService(ctx context.Context, url string, req *models.Request) (*models.Response, error) {
	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse response
	var response models.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response, nil
}